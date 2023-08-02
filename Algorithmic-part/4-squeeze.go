package main

import "fmt"

func cumPress(l, r int) int {
	if r-l >= 6 {
		// Все знают, что если сумма цифр числа делится на 9,
		// то и сумма его цифр делится на 9
		// |число делицца на девять| -> |сумма цифр делицца на 9|
		return 9
	}
	acc := 1
	for i := l; i <= r; i++ {
		acc *= i % 9
	}
	return acc % 9
}

func main() {
	var q int
	fmt.Scan(&q)

	for i := 0; i < q; i++ {
		var l, r int
		fmt.Scan(&l, &r)
		fmt.Println(cumPress(l, r))
	}
}
