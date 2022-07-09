package dget

import (
	"encoding/json"
	"github.com/gwuhaolin/lightsocks"
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

type SecureTCPConn struct {
	*lightsocks.SecureTCPConn
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
