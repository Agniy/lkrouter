package utils

import (
	"strconv"
	"time"
)

func GetNowTimestamp() string {
	return strconv.FormatInt(time.Now().Unix(), 10)
}
