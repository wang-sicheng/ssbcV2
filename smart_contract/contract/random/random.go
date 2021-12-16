package main

import (
	"math/rand"
	"time"
)

func GetRandom(args map[string]string) (interface{}, error) {
	rand.Seed(time.Now().Unix())
	return rand.Intn(100), nil
}
