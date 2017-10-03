package helpers

import "time"

func Sleep(delay time.Duration) {
	time.Sleep(delay * time.Second)
}

//CountValues: Filter an array of strings and return the number of matches and
//the len of the array
func CountValues(key string, data []string) (int, int) {
	var result int

	for _, x := range data {
		if x == key {
			result++
		}
	}
	return result, len(data)
}
