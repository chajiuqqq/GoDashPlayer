package main

import "fmt"
import "time"
import "strconv"

func maint() {

	// Use `time.Now` with `Unix` or `UnixNano` to get
	// elapsed time since the Unix epoch in seconds or
	// nanoseconds, respectively.
	now := time.Now()
	secs := now.Unix()
	nanos := now.UnixNano()
	//fmt.Println(now)

	// Note that there is no `UnixMillis`, so to get the
	// milliseconds since epoch you'll need to manually
	// divide from nanoseconds.
	millis := nanos / 10000000
	fmt.Println(secs)
	fmt.Println(millis)
	fmt.Println(millis - secs*100)

	str := strconv.FormatInt(secs, 10) + "." + strconv.FormatInt(millis-secs*100, 10)

	fmt.Println(str)

	// You can also convert integer seconds or nanoseconds
	// since the epoch into the corresponding `time`.
	//fmt.Println(time.Unix(secs, 0))
	//fmt.Println(time.Unix(0, nanos))

	//	fmt.Println(time.Now().UnixNano() / 10000000)

	//1499539382.47

	fmt.Println("**************************")

	//var f float64
	f, err := strconv.ParseFloat(str, 64) //strconv.ParseFloat("3.1415", 64)
	if err != nil {
		panic(err)
	}
	fmt.Println(strconv.FormatFloat(f, 'f', 2, 64))
}
