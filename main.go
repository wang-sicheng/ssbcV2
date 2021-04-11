package main

import (
	"bufio"
	"context"
	"flag"
	"github.com/cloudflare/cfssl/log"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
	"github.com/multiformats/go-multiaddr"
	"github.com/ssbcV2/p2p"
	"github.com/ssbcV2/util"
)

func main() {
	Start()
}

func Start() {
	//获取到执行参数中的sourcePort以及目标地址
	sourcePort := flag.Int("p", 0, "Source port number")
	dest := flag.String("d", "", "Destination multiaddr string")
	flag.Parse()

	//生成host
	host := p2p.MakeBasicHost(*sourcePort)

	//显示本节点的详情
	log.Infof("Local Node ID=%s\n", host.ID().Pretty())
	//显示本节点的公私钥
	log.Infof("Local Node Private Key=%v\nPublicKey=%v\n", util.LocalPrivateKeyStr, util.LocalPublicKeyStr)

	//为跨链准备生成本地的TCP通信server端
	go p2p.ServerSocket(host)

	if *dest == "" {
		//无目标地址
		log.Info("Waiting For Connections")
		log.Info("This Node's Multiaddresses:")
		//显示host地址
		p2p.ShowHostAddresses(host)
		host.SetStreamHandler("/p2p/1.0.0", p2p.HandleStream)
		//找一个可用的端口号
		availPort := p2p.AvailablePort(host)
		log.Infof("Run 'go run main.go -d /ip4/127.0.0.1/tcp/%v/p2p/%s' on another console.\n", availPort, host.ID().Pretty())
		select {
		//hang forever
		}
	} else {
		//有目标连接地址
		p2p.ShowHostAddresses(host)
		//将目的地址转为multiAddr
		maddr, err := multiaddr.NewMultiaddr(*dest)
		if err != nil {
			log.Error(err)
		}
		// 从multiaddr中提取info.
		info, err := peer.AddrInfoFromP2pAddr(maddr)
		if err != nil {
			log.Error(err)
		}
		// 将目标地址存储至peerStore
		// This will be used during connection and stream creation by libp2p.
		host.Peerstore().AddAddrs(info.ID, info.Addrs, peerstore.PermanentAddrTTL)
		//与目标地址建立连接
		s, err := host.NewStream(context.Background(), info.ID, "/p2p/1.0.0")
		if err != nil {
			panic(err)
		}
		host.SetStreamHandler("p2p/1.0.0", p2p.HandleStream)
		//Create a buffered stream so that read and writes are non blocking.
		rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))
		//Create a thread to read and write data.
		go p2p.WriteData(rw)
		go p2p.ReadData(rw)
		select {} // hang forever
	}
}
