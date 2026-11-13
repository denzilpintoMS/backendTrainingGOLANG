package main

import (
	"fmt"
	"strings"
)

func CheckEmail(email string) bool {
	if strings.Count(email, "@") != 1 {
		return false
	}
	atIndex := strings.Index(email, "@")
	fmt.Printf("at index @ : %v\n", atIndex)
	if atIndex <= 0 {
		return false
	}
	domain := email[atIndex+1:]
	fmt.Printf("domain : %v\n", domain)
	// slice coz . coming before @ causes a valid email to fail the case
	slicedEmail := email[atIndex:]
	fmt.Printf("sliced Email %v\n", slicedEmail)
	dotIndex := strings.Index(domain, ".")
	if dotIndex <=0 || dotIndex == len(domain)-1{
		return false
	}
	return true
}
