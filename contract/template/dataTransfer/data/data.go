package main

import (
	"github.com/ssbcV2/contract"
)

type rectangle struct {
	Height float64
	Width float64
	Annotation string
}

var Rect rectangle

func init() {
	Rect = rectangle{
		Height:     3.43,
		Width:      3.42,
		Annotation: "test",
	}
}

func GetRect(args map[string]interface{}) (interface{}, error) {
	return contract.ToBytes(Rect), nil
}
