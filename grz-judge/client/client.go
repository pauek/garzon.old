
package client

import (
	"os"
	"io"
	"bytes"
	"fmt"
	// "log"
	"bufio"
	"io/ioutil"
	"mime/multipart"
	"net/http"
)

var JudgeUrl string

func init() {
	JudgeUrl = os.Getenv("GRZ_JUDGE")
	if JudgeUrl == "" {
		JudgeUrl = "http://localhost:50000"
	}
}

func firstLine(R io.Reader) (id string, err error) {
	data := bufio.NewReader(R)
	line, _, err := data.ReadLine()
	if err != nil {
		return "", fmt.Errorf("Cannot read 1st line of resp.: %s\n", err)
	}
	return string(line), nil
}

func Submit(probid, filename string) (id string, err error) {
	var buff bytes.Buffer
	w := multipart.NewWriter(&buff)
	w.WriteField("id", probid)
	part, err := w.CreateFormFile("solution", filename)
	if err != nil {
		return "", fmt.Errorf("Cannot create form file: %s\n", err)
	}
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("Cannot read file '%s'\n", filename)
	}
	part.Write(data)
	w.Close()

	mime := fmt.Sprintf("multipart/form-data; boundary=%s", w.Boundary())
	resp, err := http.Post(JudgeUrl + "/submit/", mime, &buff)
	if err != nil {
		return "", fmt.Errorf("Cannot POST: %s\n", err)
	}
	defer resp.Body.Close()
	return firstLine(resp.Body)
}

func Status(subid string) (status string, err error) {
	url := fmt.Sprintf("%s/status/%s", JudgeUrl, subid)
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("Cannot GET: %s\n", err)
	}
	defer resp.Body.Close()
	return firstLine(resp.Body)
}