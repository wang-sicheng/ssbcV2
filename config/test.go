package config

import (
	"fmt"
	"time"
)

var a int
var b string
var c time.Time
var d bool
var e = map[string]int{}
var f = 3
var g []int

var (
	h  bool
	i  bool
	j bool
)
var k example

type example struct {
	x int
	y string
	z []int
}

func main() {
	fmt.Println("Hello")
}
