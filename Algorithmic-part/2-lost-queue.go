package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type LostQueue []string

func newLostQueue() *LostQueue {
	return new(LostQueue)
}

// Add() — добавление элемента в LostQueue
func (lq *LostQueue) Add(v string) {
	*lq = append(*lq, v)
}

// Extend() — расширение LostQueue: каждый элемент в ней дублируется
func (lq *LostQueue) Extend() {
	// nlq := newLostQueue()
	nlq := make(LostQueue, len(*lq)*2)

	j := 0
	for _, v := range *lq {
		nlq[j] = v
		nlq[j+1] = v
		j += 2
	}

	*lq = nlq
}

// Remove() — удалить элемент с начала утерянной очереди.
func (lq *LostQueue) Remove() {
	poooped := (*lq)[0]
	(*lq) = (*lq)[1:]
	fmt.Println(poooped)
}

// TODO: Переписать через мапумапу когда нибудь
/*
Поддерживаем лев гр и прав гр

правая гр тупо последнее число
левая гр: число и сколько числа осталось

(0 -> (1 -> Pair(2, 0))

2
ext
(0 -> 1)
(1 -> 2)
(0 -> (1 -> Pair(2, 0))
ext
(0 -> (1 -> Pair(4, 0))
11112222
pop
1112222
11111122222222
l = 0
11111122222222
111111222222223
r = 2
(0 -> (2 ...))
*/
func main() {
	sc := bufio.NewScanner(os.Stdin)
	// количество запросов к утерянной очереди.
	sc.Scan()
	_ = sc.Text()

	q := newLostQueue()

	for sc.Scan() {
		req := strings.Split(sc.Text(), " ")

		switch req[0] {
		case "1": // Add to LostQueue
			q.Add(req[1])
		case "2": // Extend
			q.Extend()
		case "3": //
			q.Remove()
		}
	}
}
