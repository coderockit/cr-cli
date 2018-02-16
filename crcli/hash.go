package crcli

import (
	"crypto/sha512"
	"encoding/base64"
	"strings"
)

func Hash(tohash string) string {
	//logger := loggo.GetLogger("coderockit.cli.hash")
	//[sha512.Size]byte dk = sha512.Sum512([]byte(tohash))
	//dk := make([sha512.Size]byte, sha512.Size, sha512.Size)
	dk := sha512.Sum512([]byte(tohash))
	hash := base64.StdEncoding.EncodeToString(dk[:])
	//hash := string(dk)
	HashLogger.Debugf("Hash: %s", hash)
	return hash
}

//func Hash(tohash string) string {
//	logger := loggo.GetLogger("coderockit.cli.hash")

//	salt := make([]byte, ConfInt("scrypt.saltByteArraySize", 16))
//	rand.Read(salt)

//	N := ConfInt("scrypt.memoryCostParameter", 1<<15)
//	// r and p must satisfy r * p < 2^30
//	r := ConfInt("scrypt.r", 8)
//	p := ConfInt("scrypt.p", 1)
//	keyLen := ConfInt("scrypt.keylen", 32)

//	dk, err := scrypt.Key([]byte(tohash), salt, N, r, p, keyLen)
//	if err != nil {
//		HashLogger.Criticalf("Error: %s", err)
//	}
//	hash := base64.StdEncoding.EncodeToString(dk)
//	HashLogger.Debugf("Hash: %s", hash)
//	return hash
//}

func UrlEncodeBase64(base64Str string) string {
	base64Input := base64Str
	base64Input = strings.Replace(base64Input, "+", ".", -1)
	base64Input = strings.Replace(base64Input, "/", "_", -1)
	base64Input = strings.Replace(base64Input, "=", "-", -1)
	return base64Input
}

func UrlDecodeBase64(encodedBase64Str string) string {
	encodedBase64Input := encodedBase64Str
	encodedBase64Input = strings.Replace(encodedBase64Input, ".", "+", -1)
	encodedBase64Input = strings.Replace(encodedBase64Input, "_", "/", -1)
	encodedBase64Input = strings.Replace(encodedBase64Input, "-", "=", -1)
	return encodedBase64Input
}
