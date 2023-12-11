package BasicAuthBruteForce

import (
	"crypto/tls"
	"encoding/base64"
	"log"
	"net/http"
)

const (
	authorizationHeader = "Authorization"

	userAgentHeader = "User-Agent"

	defaultAuthorizationHeader = "Basic "
)

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

	response, err := c.client.Do(req)
	if err != nil {
		log.Panic(err.Error())
	}

	defer func() {
		if err := response.Body.Close(); err != nil {
			log.Println(err)
		}
	}()

	if response.StatusCode != http.StatusUnauthorized {
		return true

	} else {
		return false
	}

	return true
}
