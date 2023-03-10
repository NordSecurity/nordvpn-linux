package slices

import "fmt"

func ExampleContains() {
	fmt.Println(Contains([]int{1, 2, 3}, 3))
	fmt.Println(Contains([]int{1, 2, 3}, 4))
	// Output: true
	// false
}

func ExampleContainsFunc() {
	fmt.Println(ContainsFunc([]int{1, 2, 3}, func(i int) bool { return i == 3 }))
	fmt.Println(ContainsFunc([]int{1, 2, 3}, func(i int) bool { return i == 4 }))
	// Output: true
	// false
}

func ExampleDelete() {
	fmt.Println(Delete([]int{1, 2, 3}, 1, 2))
	fmt.Println(Delete([]int{1, 2, 3}, 0, 1))
	// Output: [1 3]
	// [2 3]
}

func ExampleFilter() {
	fmt.Println(Filter([]int{1, 2, 3}, func(i int) bool { return i == 3 }))
	fmt.Println(Filter([]int{1, 2, 3}, func(i int) bool { return i > 0 }))
	// Output: [3]
	// [1 2 3]
}

func ExampleIndexFunc() {
	fmt.Println(IndexFunc([]int{1, 2, 3}, func(i int) bool { return i == 3 }))
	fmt.Println(IndexFunc([]int{1, 2, 3}, func(i int) bool { return i == 4 }))
	// Output: 2
	// -1
}
