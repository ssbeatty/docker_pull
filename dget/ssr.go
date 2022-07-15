package dget

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"strings"
)

type SSR struct {
	Enable     bool
	ConfigPath string
	Url        string
}

type Params struct {
	Method, Passwd, Address, Port, Obfs, ObfsParam, Protocol, ProtocolParam, Remarks, Group string
}

const ssrPrefix = "ssr://"

// base64decode for decoding url from base64 strings
func base64decode(enc string) ([]byte, error) {
	if result, errs := base64.RawURLEncoding.DecodeString(enc); errs != nil {
		return base64.StdEncoding.DecodeString(enc)
	} else {
		return result, nil
	}
}

// forceDecode for force decoding strings without any errors
func forceDecode(in string) string {
	if in != "" {
		if b, err := base64decode(in); err == nil {
			in = string(b)
		}
	}
	return in
}

// decodeURI for decode URI params once
func decodeURI(uri string) (*Params, error) {
	if !strings.HasPrefix(uri, ssrPrefix) {
		return nil, errors.New("not a valid ssr string")
	} else {
		uri = uri[len(ssrPrefix):]
	}

	b, err := base64decode(uri)
	if err != nil {
		return nil, err
	}

	s := string(b)
	c := &Params{}

	i := strings.Index(s, ":")
	if i > -1 {
		c.Address = strings.TrimSpace(s[:i])
		s = s[i+1:]
	}
	i = strings.Index(s, ":")
	if i > -1 {
		c.Port = s[:i]
		s = s[i+1:]
	}
	i = strings.Index(s, ":")
	if i > -1 {
		c.Protocol = s[:i]
		s = s[i+1:]
	}
	i = strings.Index(s, ":")
	if i > -1 {
		c.Method = s[:i]
		s = s[i+1:]
	}
	i = strings.Index(s, ":")
	if i > -1 {
		c.Obfs = s[:i]
		s = s[i+1:]
	}
	i = strings.Index(s, "/")
	if i > -1 {
		c.Passwd = strings.TrimSpace(forceDecode(s[:i]))
		s = s[i+1:]
	}

	u, err := url.Parse(s)
	if err != nil {
		return nil, err
	}
	c.ObfsParam = forceDecode(u.Query().Get("obfsparam"))
	c.ProtocolParam = forceDecode(u.Query().Get("protoparam"))
	c.Remarks = forceDecode(u.Query().Get("remarks"))
	c.Group = forceDecode(u.Query().Get("group"))

	return c, nil
}

func ParseUrlBase64(url string) (*Params, error) {
	if !strings.HasPrefix(url, ssrPrefix) {
		url = ssrPrefix + url
	}
	uri, err := decodeURI(url)
	if err != nil {
		return nil, err
	}
	return uri, nil
}

func ConvertDialerURL(params Params) (s string, err error) {
	u, err := url.Parse(fmt.Sprintf(
		"ssr://%v:%v@%v:%v",
		params.Method,
		params.Passwd,
		params.Address,
		params.Port,
	))
	if err != nil {
		return
	}
	q := u.Query()
	if len(strings.TrimSpace(params.Obfs)) <= 0 {
		params.Obfs = "plain"
	}
	if len(strings.TrimSpace(params.Protocol)) <= 0 {
		params.Protocol = "origin"
	}
	q.Set("obfs", params.Obfs)
	q.Set("obfs_param", params.ObfsParam)
	q.Set("protocol", params.Protocol)
	q.Set("protocol_param", params.ProtocolParam)
	u.RawQuery = q.Encode()
	return u.String(), nil
}
