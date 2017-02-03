package pushserver

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"golang.org/x/net/http2"
)

// Apple host locations for configuring Service.
const (
	Development     = "https://api.development.push.apple.com"
	Development2197 = "https://api.development.push.apple.com:2197"
	Production      = "https://api.push.apple.com"
	Production2197  = "https://api.push.apple.com:2197"
)

const maxPayload = 4096 // 4KB at most

// Service is the Apple Push Notification Service that you send notifications to.
type Service struct {
	Host   string
	Client *http.Client
}

// NewService creates a new service to connect to APN.
func NewService(client *http.Client, host string) *Service {
	return &Service{
		Client: client,
		Host:   host,
	}
}

// NewClient sets up an HTTP/2 client for a certificate.
func NewClient(cert tls.Certificate) (*http.Client, error) {
	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
	config.BuildNameToCertificate()
	transport := &http.Transport{TLSClientConfig: config}

	if err := http2.ConfigureTransport(transport); err != nil {
		return nil, err
	}

	return &http.Client{Transport: transport}, nil
}

// Push sends a notification and waits for a response.
func (s *Service) Push(deviceToken string, headers *Headers, payload []byte) (string, error) {
	// check payload length before even hitting Apple.
	if len(payload) > maxPayload {
		return "", errors.New("entity too large")
	}

	urlStr := fmt.Sprintf("%v/3/device/%v", s.Host, deviceToken)

	req, err := http.NewRequest("POST", urlStr, bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	headers.set(&req.Header)

	resp, err := s.Client.Do(req)

	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		return string(body), err
	}

	return resp.Status, errors.New(fmt.Sprintf("response code: %d", resp.StatusCode))
}
