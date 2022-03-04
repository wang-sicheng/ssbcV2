package client

import (
	"bytes"
	"github.com/cloudflare/cfssl/log"
	"os"
	"os/exec"
)

// 静态代码检测
func staticCodeInspection(code string) (string, error) {
	dir := "./tmp/"
	err := os.MkdirAll(dir, 0777)
	if err != nil {
		log.Error(err)
		return err.Error(), err
	}

	// 将文件夹删除
	defer os.RemoveAll(dir)

	// 创建保存文件
	destFile, err := os.Create(dir + "check.go")
	if err != nil {
		log.Error("Create failed: %s\n", err)
		return err.Error(), err
	}
	defer destFile.Close()
	_, _ = destFile.WriteString(code)

	cmd := exec.Command("gosec", "-exclude=G104", "tmp")
	var stdin, stdout, stderr bytes.Buffer
	cmd.Stdin = &stdin
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	outStr, _ := "合约存在以下问题：" + string(stdout.Bytes()), string(stderr.Bytes())
	if err != nil {
		log.Errorf("cmd.Run() failed with %s\n", err)
		return outStr, err
	}

	return "", nil
}
