package dget

import (
	"encoding/base64"
)

type Auth interface {
	ParseAuthHeader() string
}

type BasicAuth struct {
	UserName string
	PassWord string
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func (b *BasicAuth) ParseAuthHeader() string {
	return "Basic " + basicAuth(b.UserName, b.PassWord)
}
