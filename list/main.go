package main

import (
	"fmt"
)

type LinkedList struct {
	value int
	next  *LinkedList
}

func main() {
	var a, b float64
	fmt.Printf("whatever: %f\n", a+b)
	head := &LinkedList{value: 3, next: &LinkedList{value: 4}}
	fmt.Printf("&head=%p, &head.next=%p\n", head, &head.next)
}

func createList(nums []int) *LinkedList {
	var head *LinkedList

	for _, num := range nums {
		node := &LinkedList{value: num}
		if head == nil {
			head = node
		} else {
			node.next = head.next
			head.next = node
		}
	}
	return head
}
