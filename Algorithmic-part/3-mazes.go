package main

import (
	"fmt"
	"log"
	"sort"
)

func parseInput(n int) ([]int, error) {
	slice := make([]int, n)

	for i := range slice {
		_, err := fmt.Scan(&slice[i])

		if err != nil {
			return slice[:i], err
		}
	}

	return slice, nil
}

func main() {
	nmq, err := parseInput(3)
	if err != nil {
		log.Fatal(err)
	}

	n, m, q := nmq[0], nmq[1], nmq[2]

	grid := make([][]int, n)
	for i := range grid {

		grid[i], err = parseInput(m)
		if err != nil {
			log.Fatal(err)
		}
	}

	gridT := make([][]int, m)
	for i := range gridT {

		gridT[i] = make([]int, n)
		for j := range gridT[i] {
			gridT[i][j] = grid[j][i]
		}

		sort.Ints(gridT[i])
	}

	gridC := make([][]int, n)
	for i := range gridC {
		gridC[i] = make([]int, m)

		for j := range gridC[i] {
			gridC[i][j] = grid[i][j]
		}

		sort.Ints(gridC[i])
	}

	for i := 0; i < q; i++ {
		exp, err := parseInput(3)

		if err != nil {
			log.Fatal(err)
		}

		x, y, k := exp[0], exp[1], exp[2]
		result := 0

		row := gridC[x-1]
		col := gridT[y-1]
		cur := grid[x-1][y-1]

		rowL := sort.Search(len(row), func(i int) bool {
			return row[i] >= cur-k
		})

		rowR := sort.Search(len(row), func(i int) bool {
			return row[i] > cur+k
		})

		colL := sort.Search(len(col), func(i int) bool {
			return col[i] >= cur-k
		})

		colR := sort.Search(len(col), func(i int) bool {
			return col[i] > cur+k
		})

		result += rowR - rowL + colR - colL - 2

		fmt.Println(result)
	}
}
