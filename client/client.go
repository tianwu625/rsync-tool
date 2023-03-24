package client

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

type opfsAuth struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Client struct {
	Ip     string
	client *http.Client
	token  string
	auth   opfsAuth
}

/*
 * take token from client and insert into request
 */
const (
	nameAuth = "Authorization"
)

var CallError = errors.New("response status is not ok")

func (c *Client) Do(req *http.Request) (io.ReadCloser, error) {
	req.Header.Set(nameAuth, "Bearer"+c.token)

	resp, err := c.client.Do(req)
	if err != nil {
		fmt.Printf("Do %v fail %v\n", req, err)
		return nil, err
	}
	if resp.StatusCode != 200 {
		fmt.Println(resp)
		return nil, fmt.Errorf("req %v returns %v", req, resp)
	}
	return resp.Body, nil
}

func NewClient(ip, username, password string) (*Client, error) {
	c := Client{
		Ip: ip,
		client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		},
		auth: opfsAuth{
			Username: username,
			Password: password,
		},
	}
	b, err := json.Marshal(c.auth)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", "https://"+ip+"/api/v1/auth/token", bytes.NewReader(b))
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	c.token = string(body)

	return &c, nil
}
