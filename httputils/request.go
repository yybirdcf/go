package httputils

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

var (
	client = &http.Client{}
)

func request(method string, url string, body string, headers map[string]string) (string, error) {
	m := strings.ToUpper(method)
	var request *http.Request
	if m == "GET" {
		request, _ = http.NewRequest(method, url, nil)
	} else {
		b := strings.NewReader(body)
		request, _ = http.NewRequest(method, url, b)
	}

	if headers != nil {
		for k, v := range headers {
			request.Header.Set(k, v)
		}
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response, err := client.Do(request)
	if err != nil {
		return "", err
	}

	if response.StatusCode == 200 {
		body, err := ioutil.ReadAll(response.Body)
		return string(body), err
	} else {
		return response.Status, errors.New(fmt.Sprintf("response code: %d", response.StatusCode))
	}
}

//模拟发送GET请求
func HttpGet(url string, headers map[string]string) (string, error) {
	return request("GET", url, "", headers)
}

//模拟发送POST请求
func HttpPost(url string, body string, headers map[string]string) (string, error) {
	return request("POST", url, body, headers)
}

//模拟发送PUT请求
func HttpPut(url string, body string, headers map[string]string) (string, error) {
	return request("PUT", url, body, headers)
}

//模拟发送DELETE请求
func HttpDelete(url string, body string, headers map[string]string) (string, error) {
	return request("DELETE", url, body, headers)
}
