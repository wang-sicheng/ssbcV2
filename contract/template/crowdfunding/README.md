## 跨链众筹合约
* 在 A 链上部署 deposit 合约，多个用户调用 Deposit 方法转账，最后由发布者调用 End 方法结束众筹。

* 在 B 链上部署 handout 合约，调用 GetCoin 即可拿到 A 链上 deposit 合约的数据，随后可根据数据做具体业务处理。
