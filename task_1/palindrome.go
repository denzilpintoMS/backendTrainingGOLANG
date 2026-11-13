package main

import (
	"fmt"
	"unicode"
)


func CheckPalindrome(a string) bool {
	chs :=  []rune(a)
	fmt.Printf("rune : %v\n", a)
	i, j := 0, len(chs)-1
	for i < j {
		if !unicode.IsLetter(chs[i]) && !unicode.IsDigit(chs[j]) {
			i++
			continue
		}
	  if !unicode.IsLetter(chs[i])&& !unicode.IsDigit(chs[j]) {
			j--
			continue
		}
		if unicode.ToLower(chs[i]) != unicode.ToLower(chs[j]) {
			return false
		}
		i++
		j--
	}
	return true
}
// a B baa  $ bba
