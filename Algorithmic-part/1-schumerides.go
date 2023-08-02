package main

import "fmt"

func main() {
	var n int
	fmt.Scan(&n)

	var last int

	for h := 1; h <= n; h++ {
		size := 2*h - 1
		fmt.Print(powInt(size, size)*(size-1)-last, " ")
		last += powInt(size, size)
	}
}

func powInt(a, b int) int {
	return a * b
}
