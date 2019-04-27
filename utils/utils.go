package utils

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"runtime"
	"strconv"
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

// HTTPGetWithHeader get page body with header
func HTTPGetWithHeader(url string, header map[string]string) (string, error) {
	client := &http.Client{}
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	for head := range header {
		request.Header.Add(head, header[head])
	}

	response, err := client.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)

	return string(body), err
}

// FilterInvalidCharacters replace invalid filename character
func FilterInvalidCharacters(str string) string {
	return regexp.MustCompile(`[\/\\\!\:\*\?\"\<\>\|]`).ReplaceAllString(str, "_")
}

// BKDRHash64 return a BKDR hash string
func BKDRHash64(str string) string {
	s := []byte(str)
	var seed uint64 = 131 // 31 131 1313 13131 131313 etc..
	var hash uint64
	for i := 0; i < len(s); i++ {
		hash = hash*seed + uint64(s[i])
	}

	return strconv.FormatUint((hash & 0x7FFFFFFFFFFFFFFF), 16)
}

// GetTimeFormat for system
// to fix windows invalid file name
func GetTimeFormat() string {
	if runtime.GOOS == "windows" {
		return "2006-01-02 15_04_05"
	}

	return "2006-02-02 15:04:05"
}
