package common

// 接收部署智能合约的特殊地址
const ContractDeployAddress = "00000000000000000000"

// levelDB 所有账户的key （key: AccountsKey - val: state.Accounts）
const AccountsKey = "levelDBAccountsKey"

// levelDB 账户私钥的key（在用户注册时，仅存储在client中， key: 账户地址+AccountsKeyPairSuffix - val: 该账户的私钥）
const AccountsPrivateKeySuffix = "PrivateKeySuffix"

// 新注册账户的初始余额
const InitBalance = 100000
