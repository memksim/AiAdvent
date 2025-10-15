package utils

import "time"

func GetToday() string {
	return time.Now().Format(time.DateOnly)
}
