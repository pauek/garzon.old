package client

import (
	"bufio"
	"bytes"
	"code.google.com/p/go.net/websocket"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type Client struct {
	JudgeUrl   string
	AuthToken  string
	Username   string
	httpclient http.Client
}

var DefaultClient Client

type Jar struct {
	client *Client
}

func (J Jar) SetCookies(U *url.URL, cookies []*http.Cookie) {
	// TODO: Check url is judge?
	for _, c := range cookies {
		if c.Name == "Auth" {
			J.client.AuthToken = c.Value
			break
		}
	}
}

func (J Jar) Cookies(U *url.URL) []*http.Cookie {
	return []*http.Cookie{&http.Cookie{
		Name:  "Auth",
		Value: J.client.AuthToken,
	}}
}

func NewClient(url string) *Client {
	c := &Client{}
	if url == "" {
		url = os.Getenv("GRZ_JUDGE")
		if url == "" {
			url = "http://localhost:50000"
		}
	}
	c.JudgeUrl = url
	c.httpclient = http.Client{Jar: Jar{client: c}}
	return c
}

func firstLine(R io.Reader) (id string, err error) {
	data := bufio.NewReader(R)
	line, _, err := data.ReadLine()
	if err != nil {
		return "", fmt.Errorf("Cannot read 1st line of resp.: %s", err)
	}
	return string(line), nil
}

func (C *Client) Open() (isOpen bool, err error) {
	Url := fmt.Sprintf("%s/open", C.JudgeUrl)
	resp, err := C.httpclient.Get(Url)
	if err != nil {
		return false, fmt.Errorf("Cannot GET '%s': %s", Url, err)
	}
	defer resp.Body.Close()
	line, err := firstLine(resp.Body)
	if err != nil {
		return false, fmt.Errorf("Cannot read response body: %s", err)
	}
	return line == "yes", nil
}

func (C *Client) ProblemList() (ids []string, err error) {
	Url := fmt.Sprintf("%s/list", C.JudgeUrl)
	resp, err := C.httpclient.Get(Url)
	if err != nil {
		return nil, fmt.Errorf("Cannot GET '%s': %s", Url, err)
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Cannot read response body: %s", err)
	}
	ids = strings.Split(string(data), "\n")
	if len(ids) > 0 && ids[len(ids)-1] == "" {
		ids = ids[:len(ids)-1]
	}
	return ids, nil
}

func (C *Client) Login(login, passwd string) (err error) {
	Url := fmt.Sprintf("%s/login", C.JudgeUrl)
	if login == "" {
		return fmt.Errorf("User empty")
	}
	resp, err := C.httpclient.PostForm(Url, url.Values{
		"login":  {login},
		"passwd": {passwd},
	})
	if err != nil {
		return fmt.Errorf("Cannot POST to '%s': %s", Url, err)
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("Error: %s", resp.Status)
	}
	resp.Body.Close()
	if C.AuthToken == "" {
		return fmt.Errorf("Didn't receive a cookie")
	}
	return nil
}

func (C *Client) Logout(login string) (err error) {
	if login == "" {
		return fmt.Errorf("User empty")
	}
	Url := fmt.Sprintf("%s/logout", C.JudgeUrl)
	req, err := http.NewRequest("POST", Url, nil)
	if C.AuthToken != "" {
		req.AddCookie(&http.Cookie{Name: "Auth", Value: C.AuthToken})
	}

	resp, err := C.httpclient.Do(req)
	if err != nil {
		return fmt.Errorf("Cannot POST to '%s': %s", Url, err)
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("Error: %s", resp.Status)
	}
	resp.Body.Close()
	return nil
}

func (C *Client) Submit(probid, filename string, data []byte) (id string, err error) {
	var buff bytes.Buffer
	w := multipart.NewWriter(&buff)
	w.WriteField("username", C.Username)
	w.WriteField("id", probid)
	part, err := w.CreateFormFile("solution", filename)
	if err != nil {
		return "", fmt.Errorf("Cannot create form file: %s", err)
	}
	part.Write([]byte(data))
	w.Close()

	mime := fmt.Sprintf("multipart/form-data; boundary=%s", w.Boundary())
	Url, err := url.Parse(C.JudgeUrl + "/submit")
	if err != nil {
		return "", fmt.Errorf("Cannot parse url '%s': %s", C.JudgeUrl+"/submit", err)
	}
	req, err := http.NewRequest("POST", Url.String(), &buff)
	if err != nil {
		return "", fmt.Errorf("Cannot create request: %s", err)
	}
	req.Header.Set("Content-Type", mime)
	if C.AuthToken != "" {
		req.AddCookie(&http.Cookie{Name: "Auth", Value: C.AuthToken})
	}

	resp, err := C.httpclient.Do(req)
	if err != nil {
		return "", fmt.Errorf("Cannot POST: %s", err)
	}
	if resp.StatusCode == 401 {
		return "", fmt.Errorf("Unauthorized")
	}
	defer resp.Body.Close()
	return firstLine(resp.Body)
}

func (C *Client) Status(subid string, callback func(status string)) (err error) {
	orig := C.JudgeUrl
	url := fmt.Sprintf("%s/status/%s", C.JudgeUrl, subid)
	url = strings.Replace(url, "http://", "ws://", 1)

	ws, err := websocket.Dial(url, "", orig)
	if err != nil {
		return fmt.Errorf("Cannot Connect: %s", err)
	}
	for {
		var msg string
		if err := websocket.Message.Receive(ws, &msg); err != nil {
			return fmt.Errorf("Message error: %s", err)
		}
		callback(msg)
		if msg == "Resolved" {
			break
		}
	}
	return nil
}

func (C *Client) Veredict(subid string) (veredict string, err error) {
	Url := fmt.Sprintf("%s/veredict/%s", C.JudgeUrl, subid)
	req, err := http.NewRequest("GET", Url, nil)
	if C.AuthToken != "" {
		req.AddCookie(&http.Cookie{Name: "Auth", Value: C.AuthToken})
	}

	resp, err := C.httpclient.Do(req)
	if err != nil {
		return "", fmt.Errorf("Cannot GET: %s", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("Cannot read response body: %s", err)
	}
	return string(body), nil
}

// Auth

func maybeCreateDir(dir string) error {
	info, err := os.Stat(dir)
	if err == nil {
		if !info.IsDir() {
			return fmt.Errorf("'%s' exists and is not a directory", dir)
		}
	} else {
		err := os.Mkdir(dir, 0700)
		if err != nil {
			return fmt.Errorf("Cannot create directory '%s'", dir)
		}
	}
	return nil
}

func configFile(name string, createParents bool) string {
	configDir := filepath.Join(os.Getenv("HOME"), ".config")
	if createParents {
		maybeCreateDir(configDir)
	}
	garzonDir := filepath.Join(configDir, "garzon")
	if createParents {
		maybeCreateDir(garzonDir)
	}
	return filepath.Join(garzonDir, "auth")
}

func (C *Client) SaveAuthToken() error {
	tok := C.AuthToken
	filename := configFile("auth", true)
	err := ioutil.WriteFile(filename, []byte(tok), 0600)
	if err != nil {
		return fmt.Errorf("Cannot write auth token to '%s': %s", err)
	}
	return nil
}

func (C *Client) MaybeReadAuthToken() error {
	open, err := C.Open()
	if err != nil {
		return fmt.Errorf("Cannot determine is judge is open: %s", err)
	}
	if open {
		return nil
	}
	filename := configFile("auth", false)
	_, err = os.Stat(filename)
	if err != nil {
		if os.IsNotExist(err.(*os.PathError).Err) {
			return fmt.Errorf("You should login first")
		} else {
			return fmt.Errorf("grz-judge/client.MaybeReadAuthToken: %s", err)
		}
	}
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("MaybeReadAuthToken: Cannot read '%s': %s", filename, err)
	}
	C.AuthToken = string(data)
	return nil
}

func (C *Client) RemoveAuthToken() error {
	filename := configFile("auth", false)
	err := os.Remove(filename)
	if err != nil {
		return fmt.Errorf("Cannot remove '%s': %s", filename, err)
	}
	return nil
}
