package main

import (
	"fmt"
)

func main(){
	nums := []int{10, 20, 30}

	sum, average := SumAndAverage(nums)
	fmt.Printf("Sum of nums : %v and Average of nums : %v\n", sum, average)
	evens, odds := EvenOrOdd(nums)
	fmt.Printf("No of even nums : %v and number of odd nums : %v\n", evens, odds)


	// fliter out +ve transaction 
	// ts := []int{-100,0,9,9777,23123,-239828,78, -7}
	ts := []int{}
	// ts := []int{-100}
	posTransactions := FilterTransactions(ts)
	fmt.Printf("Filtered Positive transaction : %v\n",posTransactions)


	// reverse slice
	words := []string{ "abc", "efg", "jkl", "okay", "sd"}
	reversedWords := ReverseSlice(words)
	fmt.Printf(" ReversedSlice : %v\n", reversedWords)


	// remove duplicates
	input := []int{1,2,2,4,4}
	ans := removeDuplicates(input)
	fmt.Printf("Remove duplicates input ; %v and Ans : %v\n", input, ans)


	// chunk slice
	inputCS := []int{1,2,3,4,5,6}
	anss := ChunkSlice(inputCS, 2)

	fmt.Printf("Make chunks : %v and Ans : %v\n", inputCS, anss)

	// word counter 
	sen := "go by go edit copy c cring for i in range okay i okay"
	v := wordCount(sen)
	fmt.Printf("WORD COUNTER :actual sentence : %v and Ans : %v\n", sen, v)

	// status Map 
statuses := []string{
    "pending",
    "completed",
    "failed",
    "pending",
    "completed",
    "completed",
    "failed",
    "pending",
}
ansSM := CountStatus(statuses)
fmt.Printf("Statuses : %v : ordered Statuses : %v\n", statuses,ansSM)


// safe look up users
users := map[int]string{
    1: "John",
    2: "Alice",
    3: "Bob",
    4: "David",
}
	userName, exists := SafeLookupUser(users, 3)
	fmt.Printf("Username : %v Found : %v\n", userName, exists)


// Total Per user
	userIDs := []int{1, 2, 1, 3, 2, 1}
	amounts := []int{100, 200, 300, 150, 100, 50}
	ansTPU := TotalPerUser(userIDs, amounts)
	fmt.Printf("Total Per user : %v\n",ansTPU)

	// Clean user name
	usernameInput := "    Denzil    Pinto" 
	returnAns := CleanUserName(usernameInput)
	fmt.Printf("returnAns : %v\n", returnAns)

	// valid email

	email := "test@gmail.com"
	valid := CheckEmail(email)
	fmt.Println("Valid : ", valid)

	// palindrome check
	palindromeInput := "7a@A7"
	// palindromeInput := "codekkk4554edoc"

	isPalindrome := CheckPalindrome(palindromeInput)
	fmt.Printf(" Palindrome : %v || is Palindrome : %v\n", palindromeInput, isPalindrome)

	


}



