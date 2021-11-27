package commonconst

//接收部署智能合约的特殊地址
const ContractDeployAddress = "00000000000000000000"

// levelDB 所有账户的key （key: AccountsKey - val: commonconst.Accounts）
const AccountsKey = "levelDBAccountsKey"

// Accounts 是一个集合，存储了所有账户的地址，在 levelDB 里存储了 key: 账户地址 - val: 账户信息
var Accounts = map[string]struct{}{}

// Faucet 账户（用于注册账户时给新账户转账，方便测试）
const FaucetAccountAddress = "FaucetAccountAddress"

// levelDB 账户私钥的key（在用户注册时，仅存储在client中， key: 账户地址+AccountsKeyPairSuffix - val: 该账户的私钥）
const AccountsPrivateKeySuffix = "PrivateKeySuffix"
