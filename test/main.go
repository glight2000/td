package main

import (
	"fmt"
	"time"
	"flag"
)

func main() {
	name := flag.String("name", "", "Input your username")
	department := flag.String("department", "", "Input your department")
	flag.Parse()

	fmt.Printf("Hi %s from %s!\n", *name, *department)

	count := 0
	for true {
		fmt.Printf("%d\t", time.Now().Unix())
		time.Sleep(3e9)
		count++
		if count % 10 == 0 {
			fmt.Println()
		}
	}
}
