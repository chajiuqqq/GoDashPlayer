package main

import (
	"fmt"
	"regexp"
)

func main_reg() {
	value := "PT0H9M56.46"

	// First compile the delimiter expression.
	re := regexp.MustCompile(`[PTHM.S]`)

	result := re.Split(value, -1)

	for i := range result {
		fmt.Println(i, result[i])
	}
}
