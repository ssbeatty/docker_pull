package dget

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"hash"
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
	var h hash.Hash = sha256.New()
	h.Write(msg)
	return hex.EncodeToString(h.Sum(nil))
}
