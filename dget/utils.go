package dget

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"github.com/vbauerster/mpb/v7"
	"io"
	"os"
	"strings"
)

func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return true
}

func DecodeBasicAuth(authHex string) Auth {
	decodeBytes, err := base64.StdEncoding.DecodeString(authHex)
	if err != nil {
		return nil
	}

	args := strings.Split(string(decodeBytes), ":")
	if len(args) > 1 {
		return &BasicAuth{
			UserName: args[0],
			PassWord: args[1],
		}
	}

	return nil
}

func HashSha256(msg []byte) string {
	var h = sha256.New()
	h.Write(msg)
	return hex.EncodeToString(h.Sum(nil))
}

func CopyFile(dstName, srcName string, bar *mpb.Bar) (written int64, err error) {
	src, err := os.Open(srcName)
	if err != nil {
		return
	}
	defer src.Close()

	dst, err := os.OpenFile(dstName, os.O_WRONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		return
	}
	defer dst.Close()

	if bar != nil {
		return io.Copy(dst, bar.ProxyReader(src))
	}
	return io.Copy(dst, src)
}
