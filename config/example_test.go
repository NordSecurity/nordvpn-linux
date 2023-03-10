package config

import "fmt"

func ExampleField_Set() {
	var b Field[bool]
	fmt.Printf("%t\n", b.Get())
	b.Set(false)
	fmt.Printf("%t\n", b.Get())
	b.Set(true)
	fmt.Printf("%t\n", b.Get())
	// Output: false
	// false
	// true
}
