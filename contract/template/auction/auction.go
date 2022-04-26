package main

import (
	"errors"
	"github.com/cloudflare/cfssl/log"
	"github.com/ssbcV2/contract"
	"time"
)

var Beneficiary string      // 拍卖受益人
var HighestBidder string    // 当前的最高出价人
var AuctionEnd time.Time    // 结束时间
var Ended bool              // 拍卖结束标记
var Bids = map[string]int{} // 所有竞拍者的出价

func init() {
	Beneficiary = contract.Caller()              // 受益人默认为发布合约的人
	AuctionEnd = time.Now().Add(time.Minute * 2) // 合约在发布两分钟后停止出价
}

func Bid(args map[string]string) (interface{}, error) {
	if AuctionEnd.Before(time.Now()) {
		contract.Transfer(contract.Caller(), contract.Value()) // 退回转账
		log.Info("拍卖已结束")
		return nil, errors.New("拍卖已结束")
	}

	if Bids[contract.Caller()]+contract.Value() <= Bids[HighestBidder] {
		contract.Transfer(contract.Caller(), contract.Value()) // 退回转账
		log.Info("出价无效")
		return nil, errors.New("出价无效")
	}

	HighestBidder = contract.Caller()
	Bids[contract.Caller()] += contract.Value()
	return nil, nil
}

func End(args map[string]string) (interface{}, error) {
	contract.Transfer(contract.Caller(), contract.Value()) // AuctionEnd方法不接受转账，退回
	if AuctionEnd.After(time.Now()) {
		log.Info("拍卖还未结束")
		return nil, errors.New("拍卖还未结束")
	}

	if Ended {
		log.Info("重复调用ActionEnd")
		return nil, errors.New("重复调用ActionEnd")
	}
	Ended = true

	contract.Transfer(Beneficiary, Bids[HighestBidder]) // 最高出价人拍卖成功
	for bidder, amount := range Bids {
		if bidder == HighestBidder {
			continue
		}
		contract.Transfer(bidder, amount) // 其他人拍卖失败，退回资金
	}
	return nil, nil
}

// 回退函数，当没有方法匹配时执行此方法
func Fallback(args map[string]string) (interface{}, error) {
	contract.Transfer(contract.Caller(), contract.Value()) // 将转账退回
	return nil, nil
}
