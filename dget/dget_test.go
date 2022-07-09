package dget

import (
	"testing"
)

var (
	client *Client
)

func init() {
	client = NewClient(&Config{
		NeedBar: false,
	})
}

func TestDownloadWithTag(t *testing.T) {
	tag, err := client.ParseImageTag("ubuntu:20.04")
	if err != nil {
		t.Error(err)
	}
	client.DownloadDockerImage(tag, "", "")
}

func TestDownloadWithoutTag(t *testing.T) {
	tag, err := client.ParseImageTag("ubuntu")
	if err != nil {
		t.Error(err)
	}
	client.DownloadDockerImage(tag, "", "")
}

func TestDownloadDockerPackage(t *testing.T) {
	tag, err := client.ParseImageTag("ghcr.io/ssbeatty/oms/oms:v0.5.2")
	if err != nil {
		t.Error(err)
	}
	client.DownloadDockerImage(tag, "", "")
}

func TestDownloadDockerPackagePrivate(t *testing.T) {
	tag, err := client.ParseImageTag("{ YOU Private Image }")
	if err != nil {
		t.Error(err)
	}
	client.DownloadDockerImage(tag, "", "")
}
