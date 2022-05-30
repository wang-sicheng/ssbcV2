# ssbcV2

## 环境
因为后端用到了golang-plugin，Windows暂时不支持，所以需要MacOS或者Linux操作系统运行。为了避免出现兼容性问题，请尽量使用推荐的软件版本（至少保证大版本相同）。

* node（v12.22.1）

* npm（v6.14.12）

* golang（1.16.3）

* Redis（v6.2.6）

* git

* gosec
```shell
go install github.com/securego/gosec/v2/cmd/gosec@latest
```
请务必确保将 $GOPATH/bin 配置到 PATH 路径，否则无法使用 gosec

## 拉取代码并构建

* 后端代码（plugin分支）（需要把后端项目放在 GOPATH的src目录底下，我的GOPATH是 /Users/wsc/Go，所以项目目录是 /Users/wsc/Go/src/ssbcV2）

```
git clone https://github.com/wang-sicheng/ssbcV2
cd ssbcV2
git checkout plugin
go build
```

* 前端代码
```
git clone https://github.com/wang-sicheng/visual-bctt
npm install
```

## 启动
* 区块链：在后端根目录开启1个终端
```
sh start.sh
```

* 如果需要跨链数据传输，需要启动预言机，详情参考
```
https://github.com/rjkris/ssbcOracle
```

* 前端：在前端根目录开启1个终端
```
npm run dev
```

* 通过 http://localhost:9528/ 即可访问


## 其他
* 删除数据
```
sh clear.sh 
```

* 出现异常情况
    * 查看 log 目录下的日志
  
    * 删除数据并重新启动
    ```shell
    sh restart.sh
    ```
