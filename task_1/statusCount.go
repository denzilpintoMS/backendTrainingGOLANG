package main


func CountStatus(status []string) map[string]int {
	statusMap := make(map[string]int)
	for _, v := range status {
		_, ok := statusMap[v]
		if ok {
			statusMap[v]++
		}else{
			statusMap[v] = 1
		}
	}
	return statusMap
}
