package main

func SafeLookupUser(users map[int]string, userID int) (string, bool) {
	userName, ok := users[userID]
	if ok {
		return userName, true
	}else{
		return "", false
	}
}
