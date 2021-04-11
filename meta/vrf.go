package meta

type VRFResult struct {
	Result      float64  //随机数结果
	ResultIndex [32]byte //随机数结果字节数组
	PK          []byte   //公钥供Vrf验证
	Proof       []byte   //结果证明
	Msg         string   //函数输入msg
	Count       int      //参与节点数
}
