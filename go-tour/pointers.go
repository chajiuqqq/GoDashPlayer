package main

import "fmt"

func pointers() {
	var y int
	var pointerToY *int
	var pointerToPointerToInt **int

	y = 10
	pointerToY = &y
	pointerToPointerToInt = &pointerToY

	fmt.Println(y)
	fmt.Println(pointerToY)
	fmt.Println(pointerToPointerToInt)
	fmt.Println("")
	fmt.Println(&y)                     // address of y
	fmt.Println(&pointerToY)            // address of pointerToY
	fmt.Println(&pointerToPointerToInt) // address of pointerToPointerToInt

	// fmt.Println(*y) throws an error because
	// you can't redirect without an address..
	// y only has int value of 10
	fmt.Println("")
	fmt.Println("")
	fmt.Println("")
	fmt.Println(*pointerToY)            // gives the value of y
	fmt.Println(*pointerToPointerToInt) // gives the value of pointerToY which is the address of y
	fmt.Println("")
	fmt.Println(**pointerToPointerToInt) // this gives 10, because we are redirecting twice to get y
	fmt.Println("")
	if pointerToY == *pointerToPointerToInt {
		fmt.Println("they are the same!")
	}

	if pointerToY == &y {
		fmt.Println("they are the same!")
	}

	if &pointerToY == pointerToPointerToInt {
		fmt.Println("they are the same!")
	}

	if y == **pointerToPointerToInt {
		fmt.Println("they are the same!")
	}

	if pointerToY == *pointerToPointerToInt {
		fmt.Println("they are the same!")
	}

}
