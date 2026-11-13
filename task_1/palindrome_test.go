package main

import "testing"

func TestCheckPalindromeUnicode(t *testing.T) {
	if !CheckPalindrome("é") {
		t.Error("a single Unicode character should be a palindrome")
	}
}
