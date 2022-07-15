package dget

import (
	"encoding/json"
	"errors"
	"github.com/gwuhaolin/lightsocks"
	"io/fs"
	"io/ioutil"
	"net"
	"os"
	"time"
)

type LightSock struct {
	Enable     bool
	ConfigPath string
}

type LightSockConfig struct {
	ListenAddr string `json:"listen"`
	RemoteAddr string `json:"remote"`
	Password   string `json:"password"`
}

// SecureTCPConn Dial light socks remote
type SecureTCPConn struct {
	*lightsocks.SecureTCPConn
	cipher     *lightsocks.Cipher
	remoteAddr string
}

func NewSecureTCPConn(cipher *lightsocks.Cipher, remote string) *SecureTCPConn {
	return &SecureTCPConn{
		cipher:     cipher,
		remoteAddr: remote,
	}
}

func (s *SecureTCPConn) Dial(_, _ string) (net.Conn, error) {
	structRemoteAddr, err := net.ResolveTCPAddr("tcp", s.remoteAddr)
	if err != nil {
		return nil, err
	}
	tcp, err := lightsocks.DialEncryptedTCP(structRemoteAddr, s.cipher)
	if err != nil {
		return nil, err
	}

	s.SecureTCPConn = tcp

	return s, nil
}

func (s *SecureTCPConn) Read(b []byte) (n int, err error) {
	return s.DecodeRead(b)
}

func (s *SecureTCPConn) Write(b []byte) (n int, err error) {
	return s.EncodeWrite(b)
}

func (s *SecureTCPConn) LocalAddr() net.Addr {
	return s.SecureTCPConn.ReadWriteCloser.(*net.TCPConn).LocalAddr()
}

func (s *SecureTCPConn) RemoteAddr() net.Addr {
	return s.SecureTCPConn.ReadWriteCloser.(*net.TCPConn).RemoteAddr()
}

func (s *SecureTCPConn) SetDeadline(t time.Time) error {
	return s.SecureTCPConn.ReadWriteCloser.(*net.TCPConn).SetDeadline(t)
}

func (s *SecureTCPConn) SetReadDeadline(t time.Time) error {
	return s.SecureTCPConn.ReadWriteCloser.(*net.TCPConn).SetReadDeadline(t)
}

func (s *SecureTCPConn) SetWriteDeadline(t time.Time) error {
	return s.SecureTCPConn.ReadWriteCloser.(*net.TCPConn).SetWriteDeadline(t)
}

func (c *Client) readLightSockConfig() (*LightSockConfig, error) {
	conf := &LightSockConfig{}

	if _, err := os.Stat(lightSockConfigPath); !os.IsNotExist(err) {
		file, err := os.Open(lightSockConfigPath)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		err = json.NewDecoder(file).Decode(conf)
		if err != nil {
			return nil, err
		}
	}

	return conf, nil
}

func (c *Client) readSSRConfig() (*Params, error) {

	if c.config().SSR.Url != "" {
		params, err := ParseUrlBase64(c.config().SSR.Url)
		if err != nil {
			return nil, err
		} else {
			conf, _ := json.Marshal(params)
			// write to path
			defer ioutil.WriteFile(ssrConfigPath, conf, fs.ModePerm)

			return params, nil
		}
	}
	if _, err := os.Stat(ssrConfigPath); !os.IsNotExist(err) {
		params := &Params{}
		file, err := os.Open(ssrConfigPath)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		err = json.NewDecoder(file).Decode(params)
		if err != nil {
			return nil, err
		} else {
			return params, nil
		}
	}

	return nil, errors.New("config not found")
}
