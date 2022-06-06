package main

import (
	"errors"
	"github.com/ssbcV2/contract"
)

var Money map[string]int
var Publisher string		// 合约发布人
var Ended bool              // 结束标记

func init() {
	Money = map[string]int{}
	Publisher = contract.Caller()
	Ended = false
}

func Deposit(args map[string]interface{}) (interface{}, error) {
	if Ended {
		_ = contract.Transfer(contract.Caller(), contract.Value()) // 退回转账
		contract.Info("众筹已结束")
		return nil, errors.New("众筹已结束")
	}
	Money[contract.Caller()] += contract.Value()
	return nil, nil
}

func End(args map[string]interface{}) (interface{}, error) {
	_ = contract.Transfer(contract.Caller(), contract.Value()) // End方法不接受转账，退回
	if Publisher != contract.Caller() {
		contract.Info("非合约发布者，无法结束众筹")
		return nil, errors.New("非合约发布者，无法结束众筹")
	}
	Ended = true
	return nil, nil
}
