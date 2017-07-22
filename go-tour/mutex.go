package main

import (
	"fmt"
	"sync"
	"time"
)

// SafeCounter is safe to use concurrently.
type SafeCounter struct {
	v   map[string]int
	mux sync.Mutex
}

// Inc increments the counter for the given key.
func (c *SafeCounter) Inc(key string) {
	time.Sleep(2 * time.Second)
	//	c.mux.Lock()

	// Lock so only one goroutine at a time can access the map c.v.
	c.v[key]++

	//	c.mux.Unlock()

}

// Value returns the current value of the counter for the given key.
func (c *SafeCounter) Value(key string) int {
	//c.mux.Lock()
	// Lock so only one goroutine at a time can access the map c.v.
	//defer c.mux.Unlock()
	return c.v[key]
}

func mainmmmm() {
	c := SafeCounter{v: make(map[string]int)}
	//	for i := 0; i < 1; i++ {
	go c.Inc("somekey")
	//}
	fmt.Println("main sleeping for 5 sec")
	time.Sleep(5 * time.Second)
	for c.v["somekey"] != 1 {
		fmt.Println("waiting for other thread")
		time.Sleep(1 * time.Second)
	}
	fmt.Println(c.Value("somekey"))

	fmt.Printf("datetime: %s", time.Now().Format("2015-03-07 11:06:39.00 +0000 PST"))
}
