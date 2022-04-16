# ssbcV2

## 环境
因为后端用到了golang-plugin，Windows暂时不支持，所以得用Mac或者Linux环境运行

* node（v12.22.1）

* npm（v6.14.12）

* golang（1.16.3）

* git

## 拉取代码

* 后端代码（plugin分支）（需要把后端项目放在 GOPATH的src目录底下，我的GOPATH是 /Users/wsc/Go，所以项目目录是 /Users/wsc/Go/ssbcV2）

```
git clone https://github.com/wang-sicheng/ssbcV2
cd ssbcV2
git checkout plugin
```

* 前端代码
```
git clone https://github.com/wang-sicheng/visual-bctt
```

* 安装gosec

```
go install github.com/securego/gosec/v2/cmd/gosec@latest
```



## 启动
* 构建
``` 
go build
```

* 单链：在根目录开启4个终端，分别运行4个节点和1个客户端
```
./ssbcV2 N0
./ssbcV2 N1
./ssbcV2 N2
./ssbcV2 N3
./ssbcV2 client1
```
或者
```
./ssbcV2 N4
./ssbcV2 N5
./ssbcV2 N6
./ssbcV2 N7
./ssbcV2 client2
```
  
* 跨链：在根目录开启10个终端，分别运行8个节点和2个客户端
```
./ssbcV2 N0
./ssbcV2 N1
./ssbcV2 N2
./ssbcV2 N3
./ssbcV2 N4
./ssbcV2 N5
./ssbcV2 N6
./ssbcV2 N7
./ssbcV2 client1
./ssbcV2 client2
```

* 如果需要跨链数据传输，需要启动预言机，详情参考
```
https://github.com/rjkris/ssbcOracle
```

* 在根目录启动前端项目visual-bctt
```
npm run dev
```

* 通过 http://localhost:9528/ ，即可访问后端，内容都在"集成后端信息"


* 其他
```
./ssbcV2 clear  # 删除数据
```

* 出现异常情况
1. 先删除数据 ./ssbcV2 clear
2. 按顺序重启所有节点
