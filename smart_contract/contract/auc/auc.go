package main

import (
	"errors"
	"github.com/cloudflare/cfssl/log"
	"github.com/ssbcV2/smart_contract"
	"time"
)

var beneficiary string			// 拍卖受益人
var highestBidder string		// 当前的最高出价人
var auctionEnd time.Time		// 结束时间
var ended	bool				// 拍卖结束标记
var bids = map[string]int{}		// 所有竞拍者的出价

func init() {
	beneficiary = smart_contract.Caller		// 受益人默认为发布合约的人
	auctionEnd = time.Now().Add(time.Minute * 2)
}

func Bid(args map[string]string) (interface{}, error) {
	if auctionEnd.Before(time.Now()) {
		smart_contract.Transfer(smart_contract.Caller, smart_contract.Value)	// 退回转账
		log.Info("拍卖已结束")
		return nil, errors.New("拍卖已结束")
	}

	if bids[smart_contract.Caller] + smart_contract.Value <= bids[highestBidder] {
		smart_contract.Transfer(smart_contract.Caller, smart_contract.Value)	// 退回转账
		log.Info("出价无效")
		return nil, errors.New("出价无效")
	}

	highestBidder = smart_contract.Caller
	bids[smart_contract.Caller] += smart_contract.Value
	return nil, nil
}

func AuctionEnd(args map[string]string) (interface{}, error) {
	smart_contract.Transfer(smart_contract.Caller, smart_contract.Value)	// AuctionEnd方法不接受转账，退回
	if auctionEnd.After(time.Now()) {
		log.Info("拍卖还未结束")
		return nil, errors.New("拍卖还未结束")
	}

	if ended {
		log.Info("重复调用ActionEnd")
		return nil, errors.New("重复调用ActionEnd")
	}

	_, err := smart_contract.Transfer(beneficiary, bids[highestBidder])
	if err != nil {
		log.Info("拍卖异常")
		return nil, errors.New("拍卖异常")
	}
	for bidder, amount := range bids {
		if bidder == highestBidder {
			continue
		}
		smart_contract.Transfer(bidder, amount)
	}
	return nil, nil
}

