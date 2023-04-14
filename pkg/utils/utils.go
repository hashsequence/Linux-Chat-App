package utils

import (
	"string"
	"time"
	guuid "github.com/google/uuid"
)

func GetTimeStamp() string {
	return strings.ReplaceAll(time.Now().Format("20060102150405.000000.000000000"),".","")
}

func GetNewUUID() string {
	return guuid.New().String()
}