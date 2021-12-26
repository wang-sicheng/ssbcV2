package util

import (
	"crypto"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/cloudflare/cfssl/log"
	"os"
	"strconv"
)

var LocalPrivateKey *rsa.PrivateKey
var LocalPublicKey *rsa.PublicKey
var LocalPrivateKeyStr string
var LocalPublicKeyStr string

func init() {
	//生成私钥
	LocalPrivateKey, _ = rsa.GenerateKey(rand.Reader, 1024)
	//生成公钥
	LocalPublicKey = &LocalPrivateKey.PublicKey
	LocalPrivateKeyByte, _ := json.Marshal(*LocalPrivateKey)
	LocalPrivateKeyStr = string(LocalPrivateKeyByte)
	LocalPublicKeyByte, _ := json.Marshal(*LocalPublicKey)
	LocalPublicKeyStr = string(LocalPublicKeyByte)
}

//RSA秘钥生成
func RSAGenerateKeys() (*rsa.PrivateKey, *rsa.PublicKey) {
	//生成私钥
	privateKey, _ := rsa.GenerateKey(rand.Reader, 1024)
	//生成公钥
	publishKey := &privateKey.PublicKey
	return privateKey, publishKey
}

//RSA加密
func RSAEncrypt(origData []byte, publishKey *rsa.PublicKey) []byte {
	//加密
	cipherText, _ := rsa.EncryptOAEP(md5.New(), rand.Reader, publishKey, origData, nil)
	return cipherText
}

//RSA解密
func RSADecrypt(origData []byte, privateKey *rsa.PrivateKey) []byte {
	plainText, err := rsa.DecryptOAEP(md5.New(), rand.Reader, privateKey, origData, nil)
	if err != nil {
		log.Error(err)
	}
	return plainText
}

//RSA签名
func RSASign(hashed []byte, privateKey *rsa.PrivateKey) []byte {
	//签名
	opts := &rsa.PSSOptions{SaltLength: rsa.PSSSaltLengthAuto, Hash: crypto.SHA256}
	sign, err := rsa.SignPSS(rand.Reader, privateKey, crypto.SHA256, hashed, opts)
	if err != nil {
		log.Error(err)
	}
	return sign
}

//RSA验签
func RSAVerifySign(publishKey *rsa.PublicKey, hashed []byte, sign []byte) bool {
	opts := &rsa.PSSOptions{SaltLength: rsa.PSSSaltLengthAuto, Hash: crypto.SHA256}
	//验证
	e := rsa.VerifyPSS(publishKey, crypto.SHA256, hashed, sign, opts)
	if e == nil {
		log.Info("Signature Verification Succeeded")
		return true
	} else {
		log.Info("Signature Verification Failed")
		log.Error(e)
		return false
	}
}

// 生成rsa公私钥
func GetKeyPair() (prvkey, pubkey []byte) {
	// 生成私钥文件
	privateKey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		panic(err)
	}
	derStream := x509.MarshalPKCS1PrivateKey(privateKey)
	block := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: derStream,
	}
	prvkey = pem.EncodeToMemory(block)
	publicKey := &privateKey.PublicKey
	derPkix, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		panic(err)
	}
	block = &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: derPkix,
	}
	pubkey = pem.EncodeToMemory(block)
	return
}


func main() {

	//生成私钥
	priv, _ := rsa.GenerateKey(rand.Reader, 1024)
	//生产公钥
	pub := &priv.PublicKey

	//设置明文
	plaintText := []byte("hello world")

	h := md5.New()
	h.Write(plaintText)

	hashed := h.Sum(nil)

	//签名
	opts := &rsa.PSSOptions{SaltLength: rsa.PSSSaltLengthAuto, Hash: crypto.MD5}
	sing, _ := rsa.SignPSS(rand.Reader, priv, crypto.MD5, hashed, opts)

	//认证
	e := rsa.VerifyPSS(pub, crypto.MD5, hashed, sing, opts)

	if e == nil {
		log.Info("验证成功")
	}
}

// 数字签名
func RsaSignWithSha256(data []byte, keyBytes []byte) []byte {
	h := sha256.New()
	h.Write(data)
	hashed := h.Sum(nil)
	block, _ := pem.Decode(keyBytes)
	if block == nil {
		panic(errors.New("private key error"))
	}
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		log.Info("ParsePKCS8PrivateKey err", err)
		panic(err)
	}

	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hashed)
	if err != nil {
		fmt.Printf("Error from signing: %s\n", err)
		panic(err)
	}

	return signature
}

// 签名验证
func RsaVerySignWithSha256(data, signData, keyBytes []byte) bool {
	block, _ := pem.Decode(keyBytes)
	if block == nil {
		panic(errors.New("public key error"))
	}
	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		panic(err)
	}

	hashed := sha256.Sum256(data)
	err = rsa.VerifyPKCS1v15(pubKey.(*rsa.PublicKey), crypto.SHA256, hashed[:], signData)
	if err != nil {
		//panic(err)
		log.Info("验签不通过！")
		return false
	}
	return true
}

// 如果当前目录下不存在目录Keys，则创建目录，并为各个节点生成rsa公私钥
func GenRsaKeys() {
	if !FileExists("./Keys") {
		log.Info("检测到还未生成公私钥目录，正在生成公私钥 ...")
		err := os.Mkdir("Keys", 0777)
		if err != nil {
			log.Error()
		}
		for i := 0; i <= 4; i++ {
			if !FileExists("./Keys/N" + strconv.Itoa(i)) {
				err := os.Mkdir("./Keys/N"+strconv.Itoa(i), 0777)
				if err != nil {
					log.Error()
				}
			}
			priv, pub := GetKeyPair()
			privFileName := "Keys/N" + strconv.Itoa(i) + "/N" + strconv.Itoa(i) + "_RSA_PIV"
			file, err := os.OpenFile(privFileName, os.O_RDWR|os.O_CREATE, 0777)
			if err != nil {
				log.Error(err)
			}
			defer file.Close()
			file.Write(priv)

			pubFileName := "Keys/N" + strconv.Itoa(i) + "/N" + strconv.Itoa(i) + "_RSA_PUB"
			file2, err := os.OpenFile(pubFileName, os.O_RDWR|os.O_CREATE, 0777)
			if err != nil {
				log.Error(err)
			}
			defer file2.Close()
			file2.Write(pub)
		}
		log.Info("已为节点们生成RSA公私钥")
	}
}
