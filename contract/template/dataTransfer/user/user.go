package main

import (
	"encoding/json"
	"github.com/ssbcV2/contract"
)

type rectangle struct {
	Height float64
	Width float64
	Annotation string
}

var Rect rectangle

func Area(args map[string]interface{}) (interface{}, error) {
	data, _ := contract.Call("data", "GetRect", nil)

	err := json.Unmarshal(data.([]byte), &Rect)
	if err != nil {
		contract.Info("json.Unmarshal error", err)
	}

	contract.Info(Rect)
	return nil, nil
}
