// Grade calculator: Given a score from 0 to 100, return a grade (A, B, C, D, F) using switch.

package main

import "fmt"

func gradeCalc(num int) rune {
	switch{
	case num > 100:
		fmt.Println("Out of range value")
		return 0
	case num < 0:
		fmt.Println("Out of range or invalid num")
		return 0
	case num >= 85: // Comma-separated values
		return 'A'
	case num >= 75:
		return 'B'
	case num >= 65:
		return 'C'
	case num >= 55:
		return 'D'
	default:
		return 'F'
	}
}


