package meta

import "crypto/rsa"

//type TSS struct {
//	FinalSignature []byte   //最终签名
//	Submitter      []byte   //提交者
//	Shares         [][]byte //达到门限阈值时参与签名者的集合
//}

type ProposalSign struct {
	Hash      []byte        //需要签名的信息
	PubKey    rsa.PublicKey //公钥信息--进行验签
	Sign      []byte        //签名
	Threshold int           //需要满足的阈值
}
