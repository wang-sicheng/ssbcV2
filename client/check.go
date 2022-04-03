package client

import (
	"bytes"
	"fmt"
	"github.com/cloudflare/cfssl/log"
	"github.com/ssbcV2/config"
	"os"
	"os/exec"
)

func check(code string) (string, error) {
	result, err := createFile(code)
	if err != nil {
		return result, err
	}
	defer deleteFile()

	// 静态代码检测
	if config.Get("check.code_check").(bool) {
		result, err := codeCheck()
		if err != nil {
			return result, err
		}
	}

	// 模型检测
	if config.Get("check.model_check").(bool) {
		result, err := modelCheck()
		if err != nil {
			return result, err
		}
	}
	return "", nil
}

// 静态代码检测
func codeCheck() (string, error) {
	cmd := exec.Command("gosec", "-exclude=G104,G404", "tmp")
	var stdin, stdout, stderr bytes.Buffer
	cmd.Stdin = &stdin
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	outStr, _ := "合约存在以下问题："+string(stdout.Bytes()), string(stderr.Bytes())
	if err != nil {
		log.Errorf("cmd.Run() failed with %s\n", err)
		return outStr, err
	}
	return "", nil
}

// 模型检测
func modelCheck() (string, error) {
	cmd := exec.Command("gomela", "fs", "tmp")
	var stdin, stdout, stderr bytes.Buffer
	cmd.Stdin = &stdin
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	outStr, _ := string(stdout.Bytes()), string(stderr.Bytes())
	if err != nil {
		log.Errorf("cmd.Run() failed with %s\n", err)
		return outStr, err
	}
	fmt.Println("合约模型检测结果：\n" + outStr)
	return "", nil
}

// 创建临时文件保存合约，用于代码检测和模型检测
func createFile(code string) (string, error) {
	dir := "./tmp/"
	err := os.MkdirAll(dir, 0777)
	if err != nil {
		log.Error(err)
		return err.Error(), err
	}

	// 创建保存文件
	destFile, err := os.Create(dir + "tmp.go")
	if err != nil {
		log.Error("Create failed: %s\n", err)
		return err.Error(), err
	}
	defer destFile.Close()
	_, _ = destFile.WriteString(code)
	return "", nil
}

func deleteFile() {
	dir := "./tmp"
	// 将文件夹删除
	os.RemoveAll(dir)
}
