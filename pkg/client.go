package BasicAuthBruteForce

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
)

const (
	authorizationHeader = "Authorization"

	userAgentHeader = "User-Agent"

	defaultAuthorizationHeader = "Basic "
)

var maxRetries int = 20
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

func (c *Client) SetHeader(siteurl, agent, user, pass string) bool {

	req, err := http.NewRequest(http.MethodGet, siteurl, nil)
	if err != nil {
		log.Panic(err.Error())
	}

	// Encode to base64
	data := user + ":" + pass
	userpass := base64.StdEncoding.EncodeToString([]byte(data))

	// Set HTTP headers

	req.Header.Set(userAgentHeader, agent)

	req.Header.Set(authorizationHeader, defaultAuthorizationHeader+userpass)
	for i := 0; i < maxRetries; i++ {
		response, err = c.client.Do(req)
		if err == nil {
			// Success, break out of the loop
			break
		} else {

			fmt.Println(err.Error(), i)
			err = nil
		}
	}
	defer func() {
		if err := response.Body.Close(); err != nil {
			recover()
			log.Println(err)
		}
	}()

	if response.StatusCode == http.StatusOK {
		return true

	} else if response.StatusCode == http.StatusUnauthorized {

		return false
	} else {
		fmt.Println(response.StatusCode)
		return false
	}

	return true
}
