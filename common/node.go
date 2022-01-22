package common

const Client1ToNodeAddr = "127.0.0.1:8008" // 客户端与节点通信的监听地址
const Client1ToUserAddr = ":8009"          // 客户端与前端用户通信监听地址
const Client2ToNodeAddr = "127.0.0.1:8010"
const Client2ToUserAddr = ":8011"

// 节点个数
const NodeCount = 4

var Ssbc1Nodes = []string{"N0", "N1", "N2", "N3", "client1"}
var Ssbc2Nodes = []string{"N4", "N5", "N6", "N7", "client2"}

var NodeTable1 = map[string]string{
	"N0": "127.0.0.1:8000",
	"N1": "127.0.0.1:8001",
	"N2": "127.0.0.1:8002",
	"N3": "127.0.0.1:8003",
}

var NodeTable2 = map[string]string{
	"N4": "127.0.0.1:8004",
	"N5": "127.0.0.1:8005",
	"N6": "127.0.0.1:8006",
	"N7": "127.0.0.1:8007",
}
