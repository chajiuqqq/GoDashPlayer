package main

import "fmt"
import "strings"

//DashPlayback{} //Notice the DashPlayback{}. The new(DashPlayback) was just a syntactic sugar for &DashPlayback{} and we don't need a pointer to the Foo

type Person struct {
	firstName string
	lastName  string
	number    int
	car       *Car
}

type Car struct {
	brand string
}

func main3() {

	adr := "https://caddy.quic/BigBuckBunny_4s.mpd"
	splited := strings.Split(adr, "")
	fmt.Println(splited[0])
	fmt.Println(splited[1])
	fmt.Println(splited[3])

	person := Person{
		firstName: "Alice",
		lastName:  "Dow",
		number:    11,
	}

	person1 := Person{}
	//person1 := new(Person)

	person2 := Person{
		firstName: "Irmak",
		lastName:  "Ozturk",
		number:    33,
	}

	person3 := Person{
		firstName: "Sevket",
		lastName:  "Arisu",
		number:    99,
	}
	car1 := Car{}
	car1.brand = "Honda"

	car2 := Car{}
	car2.brand = "Jeep"

	car3 := new(Car) //  &Car{} ile ayni sey
	car3.brand = "Audi"

	person1, personnumber := modifyPerson(person1) //atama yaptigimiz icin ismi  degistirdi
	fmt.Println("personnumber", personnumber)

	//changeNameByValue(person)
	//changeNameByPointer(&person)
	//person.firstName = getName()

	//person.firstName = getNameFromPerson(&person2)

	//person.firstName, person.lastName = getNameSurnameFromPerson(&person2)

	//person.firstName, person.lastName = getNameSurnameFromPerson2(&person2)

	//	person.firstName, person.lastName, person.number = getNameSurnameFromPerson3(&person2)

	//scopyFirstName(&person, &person2)
	buyCar(&person, car1)
	buyCar2(&person2, &car2)
	buyCar2(&person3, car3)

	buyCar3(&person, *car3) //car3=&Car{}  yukarda

	fmt.Println("person:", person)
	fmt.Println("person.car:", *person.car) //print etmek için * kullanıldı
	fmt.Println("person2:", person2)
	fmt.Println("person2.car:", *person2.car) //print etmek için * kullanıldı
	fmt.Println("person3:", person3)
	fmt.Println("person3.car:", *person3.car) //print etmek için * kullanıldı
}

func modifyPerson(p Person) (Person, int) {
	p.firstName = "AAA"
	p.lastName = "BBB"
	return p, p.number
}

func buyCar(p *Person, c Car) {
	p.car = &c
}

func buyCar2(p *Person, c *Car) {
	p.car = c
}
func buyCar3(p *Person, c Car) {
	p.car = &c
}

func changeNameByValue(p Person) {
	p.firstName = "Bob"
}

func changeNameByPointer(p *Person) {
	p.firstName = "Bob"
}

func copyFirstName(p1 *Person, p2 *Person) {
	p2.firstName = p1.firstName
}

func getName() string {
	return "Sevket"
}

func getNameFromPerson(p *Person) string {
	return p.firstName
}

func getNameSurnameFromPerson(p *Person) (string, string) {
	return p.firstName, p.lastName
}

func getNameSurnameFromPerson2(p *Person) (string, string) {
	p.number = 44
	return p.firstName, p.lastName
}
func getNameSurnameFromPerson3(p *Person) (string, string, int) {
	p.number = 55
	return p.firstName, p.lastName, p.number
}
