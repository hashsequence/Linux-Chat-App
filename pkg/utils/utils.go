package utils

import (
	"strings"
	"time"

	guuid "github.com/google/uuid"
)

func GetTimeStamp() string {
	return strings.ReplaceAll(time.Now().Format("20060102150405.000000.000000000"), ".", "")
}

func GetNewUUID() string {
	return guuid.New().String()
}

func GetArgsArr(s string) []string {
	arr := []string{}
	currStr := []byte{}
	for i := 0; i < len(s); i++ {
		if s[i] == ' ' {
			arr = append(arr, string(currStr))
			currStr = []byte{}
		} else {
			currStr = append(currStr, s[i])
		}
	}
	if len(currStr) > 0 {
		arr = append(arr, string(currStr))
	}
	return arr
}
