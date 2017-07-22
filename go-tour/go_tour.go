package main

import (
	"fmt"
)

type User struct {
	Name string
}

/*func main() {
	u := &User{Name: "Leto"} //define u as pointer and assing the adress of the object that has Name=Leto
	println(u)               //print  value of u (adress)
	println(u.Name)          //print value of u.Name (string)
	Modify(u)
	println(u.Name)
	//pointers()
}

func Modify(u *User) {
	u = &User{Name: "Paul"}
}*/

/*func main() {
	u := User{Name: "Leto"}
	println(u.Name)
	Modify(u)
	println(u.Name)
}

func Modify(u User) {
	u.Name = "Duncan"
}*/

func main2() {
	fmt.Println("start")
	u := new(User)
	//	u = &User{Name: "Leto"}
	fmt.Println(u.Name)
	Modify(&u)
	fmt.Println(u.Name)
}

func Modify(u **User) {
	*u = &User{Name: "Bob"}
}
