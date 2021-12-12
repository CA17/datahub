package common

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"hash/fnv"
	"os"

	// jsoniter "github.com/json-iterator/go"

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
