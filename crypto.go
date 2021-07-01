package utils

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"io"
	"os"
)

// SHA1Sum 计算字符串
func SHA1(str string) string {
	r := sha1.Sum([]byte(str))
	return hex.EncodeToString(r[:])
}

// MD5 计算字符串
func MD5(str string) string {
	r := md5.Sum([]byte(str))
	return hex.EncodeToString(r[:])
}

// SHA1Sum 计算文件
func SHA1Sum(path string) (string, error) {
	file, err := os.Open(path)
	defer file.Close()

	if err != nil {
		return "", err
	}

	h := sha1.New()
	_, err = io.Copy(h, file)

	if err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

// SHA1Sum 计算文件 计算文件
func MD5Sum(path string) (string, error) {
	file, err := os.Open(path)
	defer file.Close()

	if err != nil {
		return "", err
	}

	h := md5.New()
	_, err = io.Copy(h, file)

	if err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}
