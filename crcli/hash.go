package crcli

import (
	"encoding/base64"
	"math/rand"

	"github.com/juju/loggo"
	"golang.org/x/crypto/scrypt"
)

func Hash(tohash string) string {
	logger := loggo.GetLogger("coderockit.cli.hash")

	salt := make([]byte, ConfInt("scrypt.saltByteArraySize", 16))
	rand.Read(salt)

	N := ConfInt("scrypt.memoryCostParameter", 1<<15)
	// r and p must satisfy r * p < 2^30
	r := ConfInt("scrypt.r", 8)
	p := ConfInt("scrypt.p", 1)
	keyLen := ConfInt("scrypt.keylen", 32)

	dk, err := scrypt.Key([]byte(tohash), salt, N, r, p, keyLen)
	if err != nil {
		logger.Criticalf("Error: %s", err)
	}
	hash := base64.StdEncoding.EncodeToString(dk)
	logger.Debugf("Hash: %s", hash)
	return hash
}
