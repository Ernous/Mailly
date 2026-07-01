package api

import (
	"fmt"

	"github.com/emersion/go-sasl"
)

type SASLClient struct {
	username    string
	accessToken string
	mechanism   string
}

func NewSASLClient(username, accessToken, mechanism string) sasl.Client {
	return &SASLClient{
		username:    username,
		accessToken: accessToken,
		mechanism:   mechanism,
	}
}

func (c *SASLClient) Start() (mech string, ir []byte, err error) {
	if c.mechanism == "OAUTHBEARER" {
		ir = []byte(fmt.Sprintf("n,user=%s,\x01auth=Bearer %s\x01\x01", c.username, c.accessToken))
	} else {
		ir = []byte(fmt.Sprintf("user=%s\x01auth=Bearer %s\x01\x01", c.username, c.accessToken))
	}
	return c.mechanism, ir, nil
}

func (c *SASLClient) Next(challenge []byte) ([]byte, error) {
	// For OAUTHBEARER: if server sends an error JSON in the challenge,
	// respond with empty string to abort gracefully (RFC 7628 §3.2.3)
	if c.mechanism == "OAUTHBEARER" {
		return []byte(""), nil
	}
	return []byte(fmt.Sprintf("user=%s\x01auth=Bearer %s\x01\x01", c.username, c.accessToken)), nil
}
