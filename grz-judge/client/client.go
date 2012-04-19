package client

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
)

var JudgeUrl string
var AuthToken string

var client http.Client

type Jar struct{}

func (J Jar) SetCookies(U *url.URL, cookies []*http.Cookie) {
	// TODO: Check url is judge?
	for _, c := range cookies {
		if c.Name == "Auth" {
			AuthToken = c.Value
			break
		}
	}
}

func (J Jar) Cookies(U *url.URL) []*http.Cookie {
	return []*http.Cookie{&http.Cookie{Name: "Auth", Value: AuthToken}}
}

func init() {
	JudgeUrl = os.Getenv("GRZ_JUDGE")
	if JudgeUrl == "" {
		JudgeUrl = "http://localhost:50000"
	}
	client.Jar = Jar{}
}

func firstLine(R io.Reader) (id string, err error) {
	data := bufio.NewReader(R)
	line, _, err := data.ReadLine()
	if err != nil {
		return "", fmt.Errorf("Cannot read 1st line of resp.: %s", err)
	}
	return string(line), nil
}

func Login(login, passwd string) (err error) {
	Url := fmt.Sprintf("%s/login", JudgeUrl)
	if login == "" {
		return fmt.Errorf("User empty")
	}
	resp, err := client.PostForm(Url, url.Values{
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
	if AuthToken == "" {
		return fmt.Errorf("Didn't receive a cookie")
	}
	return nil
}

func Logout(login string) (err error) {
	if login == "" {
		return fmt.Errorf("User empty")
	}
	Url := fmt.Sprintf("%s/logout", JudgeUrl)
	req, err := http.NewRequest("POST", Url, nil)
	req.AddCookie(&http.Cookie{Name: "Auth", Value: AuthToken})

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Cannot POST to '%s': %s", Url, err)
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("Error: %s", resp.Status)
	}
	resp.Body.Close()
	return nil
}


func Submit(probid, filename string) (id string, err error) {
	var buff bytes.Buffer
	w := multipart.NewWriter(&buff)
	w.WriteField("id", probid)
	part, err := w.CreateFormFile("solution", filename)
	if err != nil {
		return "", fmt.Errorf("Cannot create form file: %s", err)
	}
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("Cannot read file '%s'", filename)
	}
	part.Write(data)
	w.Close()

	mime := fmt.Sprintf("multipart/form-data; boundary=%s", w.Boundary())
	Url, err := url.Parse(JudgeUrl + "/submit")
	if err != nil {
		return "", fmt.Errorf("Cannot parse url '%s': %s", JudgeUrl+"/submit", err)
	}
	req, err := http.NewRequest("POST", Url.String(), &buff)
	if err != nil {
		return "", fmt.Errorf("Cannot create request: %s", err)
	}
	req.Header.Set("Content-Type", mime)
	req.AddCookie(&http.Cookie{Name: "Auth", Value: AuthToken})

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("Cannot POST: %s", err)
	}
	if resp.StatusCode == 401 {
		return "", fmt.Errorf("Unauthorized")
	}
	defer resp.Body.Close()
	return firstLine(resp.Body)
}

func Status(subid string) (status string, err error) {
	Url := fmt.Sprintf("%s/status/%s", JudgeUrl, subid)
	req, err := http.NewRequest("GET", Url, nil)
	req.AddCookie(&http.Cookie{Name: "Auth", Value: AuthToken})

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("Cannot GET: %s", err)
	}
	switch {
	case resp.StatusCode == 401:
		return "", fmt.Errorf("Unauthorized")
	case resp.StatusCode != 200:
		return "", fmt.Errorf("Wrong status: %s", err)
	}
	defer resp.Body.Close()
	return firstLine(resp.Body)
}

func Veredict(subid string) (veredict string, err error) {
	Url := fmt.Sprintf("%s/veredict/%s", JudgeUrl, subid)
	req, err := http.NewRequest("GET", Url, nil)
	req.AddCookie(&http.Cookie{Name: "Auth", Value: AuthToken})

	resp, err := client.Do(req)
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
