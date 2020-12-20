///usr/bin/env yaegi run "$0" "$@"; exit
package main

import "fmt"
import "flag"

func main() {
	var str_val = ""
	flag.StringVar(&str_val, "s", "World", "provide a string")
	var n = flag.Int("i", 1, "provide an int")
	flag.Parse()

	fmt.Println("Int flag:", *n)
	fmt.Println("Flag: Hello", str_val)
}
