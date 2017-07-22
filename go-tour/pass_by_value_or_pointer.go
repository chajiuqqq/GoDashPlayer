//http://goinbigdata.com/golang-pass-by-pointer-vs-pass-by-value/

//Passing by value
package main

import "fmt"

type Person2 struct {
	firstName string
	lastName  string
}

func changeName(p Person2) {
	p.firstName = "Bob"
}

func mainpp() {
	person := Person2{
		firstName: "Alice",
		lastName:  "Dow",
	}

	changeName(person)

	fmt.Println(person)
}

//Passing by pointer
/*package main

import "fmt"

type Person struct {
    firstName string
    lastName  string
}

func changeName(p *Person) {
    p.firstName = "Bob"
}

func main() {
    person := Person {
        firstName: "Alice",
        lastName: "Dow",
    }

    changeName(&person)

    fmt.Println(person)
}*/
