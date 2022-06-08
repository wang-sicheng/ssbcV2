package main

import (
	"errors"
	"github.com/ssbcV2/contract" // 调用其他智能合约时引入
	"github.com/ssbcV2/meta"
	"strconv"
	"time"
)

var AvgConsensusTime time.Duration
var AvgRequestTime time.Duration
var Amount int64

type MonitorInfo struct {
	StartConsensusTime string
	ConsensusCostTime int
	DataRequestTime int
	Data string
	ConsensusResult bool
}
var Report MonitorInfo

func init() {
	Amount = 0
}

func CallbackDataMonitor(args map[string]interface{}) (interface{}, error) {
	StartConsensusTime, ok := args["StartConsensusTime"].(string)
	if !ok {
		contract.Info("miss StartConsensusTime field")
		return nil, errors.New("miss StartConsensusTime field")
	}
	ConsensusCostTimeStr, ok := args["ConsensusCostTime"].(string)
	ConsensusCostTime, _ := strconv.Atoi(ConsensusCostTimeStr)
	if !ok {
		contract.Info("miss ConsensusCostTime field")
		return nil, errors.New("miss ConsensusCostTime field")
	}
	DataRequestTimeStr, ok := args["DataRequestTime"].(string)
	DataRequestTime, _ := strconv.Atoi(DataRequestTimeStr)
	if !ok {
		contract.Info("miss DataRequestTime field")
		return nil, errors.New("miss DataRequestTime field")
	}
	Data, ok := args["Data"].(string)
	if !ok {
		contract.Info("miss Data field")
		return nil, errors.New("miss Data field")
	}
	ConsensusResult, ok := args["ConsensusResult"].(bool)
	if !ok {
		contract.Info("miss ConsensusResult field")
		return nil, errors.New("miss ConsensusResult field")
	}
	// 更新report
	Report.StartConsensusTime = StartConsensusTime
	Report.ConsensusCostTime = ConsensusCostTime
	Report.DataRequestTime = DataRequestTime
	Report.Data = Data
	Report.ConsensusResult = ConsensusResult
	contract.Info("链下报告更新成功")
	// 更新统计值
	AvgConsensusTime = time.Duration((int64(AvgConsensusTime)*Amount + int64(Report.ConsensusCostTime)) / (Amount + 1))
	AvgRequestTime = time.Duration((int64(AvgRequestTime)*Amount + int64(Report.DataRequestTime)) / (Amount + 1))
	Amount += 1
	contract.Info("统计值更新成功")
	return nil, nil
}


// 回退函数，当没有方法匹配时执行此方法
func Fallback(args map[string]interface{}) (interface{}, error) {
	return meta.ContractUpdateData{}, nil
}
