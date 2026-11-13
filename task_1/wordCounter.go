package main

import (
	"fmt"
	"strings"
)

func wordCount(sentence string) map[string]int {
	input := strings.Fields(sentence)
	fmt.Println("input : ", input)
	counter := make(map[string]int)
	for _, s := range input {
		_, ok := counter[s]
		if ok {
			counter[s]++
		}else{
			counter[s] = 1
		}
	}
	return counter
}
