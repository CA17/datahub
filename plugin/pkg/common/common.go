package common

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"hash/fnv"
	"os"
	"time"

	// jsoniter "github.com/json-iterator/go"

	"github.com/form3tech-oss/jwt-go"
)

// var json = jsoniter.ConfigCompatibleWithStandardLibrary

func FileExists(file string) bool {
	info, err := os.Stat(file)
	return err == nil && !info.IsDir()
}

func IsFilePath(file string) bool {
	info, err := os.Stat(file)
	return err == nil && !info.IsDir()
}

func Md5Hash(src string) string {
	hash := md5.Sum([]byte(src))
	return hex.EncodeToString(hash[:])
}

func StringHash(str string) uint64 {
	h := fnv.New64a()
	_, err := h.Write([]byte(str))
	if err != nil {
		panic(err)
	}
	return h.Sum64()
}

func ToJson(v interface{}) string {
	bs, _ := json.MarshalIndent(v, "", "  ")
	return string(bs)
}

func CreateToken(secret string) (string, error) {
	// Create token
	token := jwt.New(jwt.SigningMethodHS256)
	// Set claims
	claims := token.Claims.(jwt.MapClaims)
	claims["usr"] = "teamsdns"
	claims["uid"] = "teamsdns"
	claims["lvl"] = "api"
	claims["exp"] = time.Now().Add(time.Hour).Unix()
	return token.SignedString([]byte(secret))
}
