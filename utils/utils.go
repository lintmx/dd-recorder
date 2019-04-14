package utils

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"net/http"
)

// GetMd5 get md5
func GetMd5(s string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(s)))
}

// HTTPGet get page body
func HTTPGet(url string) (string, error) {
	response, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)

	return string(body), err
}
