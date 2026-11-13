package main

import (
	"strings"
)


func CleanUserName(userName string) string {
	ans := strings.TrimSpace(userName)
	ans = strings.ReplaceAll(ans, " ", "_")
	ans = strings.ToLower(ans)
	return ans
}
