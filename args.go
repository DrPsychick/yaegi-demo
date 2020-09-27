#!/usr/bin/env yaegi
package main

import "fmt"
import "flag"

func main() {
	var string = ""
	flag.StringVar(&string, "s", "World", "provide a string")
	var n = flag.Int("i", 1, "provide an int")
	flag.Parse()

	fmt.Println("Int flag: ", *n)
	fmt.Println("Flag: Hello ", string)
}

