package BasicAuthBruteForce

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"golang.org/x/net/proxy"
	"log"
	"net/http"
)

//var count int

const (
	authorizationHeader = "Authorization"

	userAgentHeader = "User-Agent"

	defaultAuthorizationHeader = "Basic "
)

var maxRetries = 20
var response *http.Response

type Client struct {
	client *http.Client
}

func NewClient() *Client {
	return &Client{
		client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
	}
}
func (c *Client) SetProxy(dialer proxy.Dialer) {
	transport := &http.Transport{
		Dial: dialer.Dial,
	}
	c.client.Transport = transport
}
func (c *Client) SetHeader(siteurl, agent, user, pass string) bool {

	req, err := http.NewRequest(http.MethodGet, siteurl, nil)
	if err != nil {
		log.Panic(err.Error())
	}

	// Encode to base64
	data := user + ":" + pass
	userpass := base64.StdEncoding.EncodeToString([]byte(data))

	//	fmt.Println(data)
	// Set HTTP headers

	req.Header.Set(userAgentHeader, agent)

	req.Header.Set(authorizationHeader, defaultAuthorizationHeader+userpass)
	for i := 0; i < maxRetries; i++ {
		response, err = c.client.Do(req)
		if err == nil && response != nil {
			// Success, break out of the loop
			break
		} else {

		}
	}

	defer func() {
		if err := response.Body.Close(); err != nil {
			recover()
			log.Println(err)
		}
	}()

	if response.StatusCode == http.StatusOK || response.StatusCode == http.StatusMovedPermanently {

		return true

	} else if response.StatusCode == http.StatusUnauthorized {
		//fmt.Println(response.StatusCode, user, pass)
		return false
	} else {
		fmt.Println(response.StatusCode, "User name :"+user, "Password : "+pass)
		return false
	}

	return true
}
