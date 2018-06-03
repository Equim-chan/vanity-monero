package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

var stdin = bufio.NewScanner(os.Stdin)

func prompt(question string) string {
	for {
		fmt.Print(question + " ")
		stdin.Scan()
		ans := strings.TrimSpace(stdin.Text())
		if ans != "" {
			return ans
		}
		fmt.Println("can't be empty")
	}
}

func promptComfirm(question string) bool {
	return prompt(question) == "y"
}

func promptNumber(question string, min, max int) int {
	for {
		n, err := strconv.Atoi(prompt(question))
		switch {
		case err != nil:
			fmt.Println("invalid number")
		case n < min || n > max:
			fmt.Println("invalid range")
		default:
			return n
		}
	}
}
