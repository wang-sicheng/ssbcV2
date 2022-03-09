package main

import (
	"errors"
	"github.com/cloudflare/cfssl/log"
	"github.com/ssbcV2/contract"
	"time"
)

var beneficiary string      // 拍卖受益人
var highestBidder string    // 当前的最高出价人
var auctionEnd time.Time    // 结束时间
var ended bool              // 拍卖结束标记
var bids = map[string]int{} // 所有竞拍者的出价

func init() {
	beneficiary = contract.Caller()              // 受益人默认为发布合约的人
	auctionEnd = time.Now().Add(time.Minute * 2) // 合约在发布两分钟后停止出价
}

func Bid(args map[string]string) (interface{}, error) {
	if auctionEnd.Before(time.Now()) {
		contract.Transfer(contract.Caller(), contract.Value()) // 退回转账
		log.Info("拍卖已结束")
		return nil, errors.New("拍卖已结束")
	}

	if bids[contract.Caller()]+contract.Value() <= bids[highestBidder] {
		contract.Transfer(contract.Caller(), contract.Value()) // 退回转账
		log.Info("出价无效")
		return nil, errors.New("出价无效")
	}

	highestBidder = contract.Caller()
	bids[contract.Caller()] += contract.Value()
	return nil, nil
}

func End(args map[string]string) (interface{}, error) {
	contract.Transfer(contract.Caller(), contract.Value()) // AuctionEnd方法不接受转账，退回
	if auctionEnd.After(time.Now()) {
		log.Info("拍卖还未结束")
		return nil, errors.New("拍卖还未结束")
	}

	if ended {
		log.Info("重复调用ActionEnd")
		return nil, errors.New("重复调用ActionEnd")
	}
	ended = true

	contract.Transfer(beneficiary, bids[highestBidder]) // 最高出价人拍卖成功
	for bidder, amount := range bids {
		if bidder == highestBidder {
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
