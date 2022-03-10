package contract

// 合约调用上下文
type context struct {
	Name    string            // 当前执行的合约的名称
	Address string 			  // 合约地址
	Method  string            // 被调用的方法
	Args    map[string]string // 参数
	Balance int               // 合约账户的余额
	Caller  string            // 调用者地址（合约账户、外部账户）
	Origin  string            // 最初调用者（外部账户），如果不涉及合约调用合约，那么 Caller == Origin
	Value   int               // 调用合约时的转账金额
}

// 合约调用栈
type contextStack struct {
	contexts []context
}

var stack contextStack // 合约调用栈
var curContext context // 当前合约调用的context

func init() {
	curContext = context{}
	stack = contextStack{[]context{}}
}

func (t *contextStack) Push(c context) {
	t.contexts = append(t.contexts, c)
}

func (t *contextStack) Pop() context {
	if !t.IsEmpty() {
		top := t.contexts[len(t.contexts)-1]
		t.contexts = t.contexts[:len(t.contexts)-1]
		return top
	}
	return context{}
}

func (t *contextStack) Top() context {
	if !t.IsEmpty() {
		return t.contexts[len(t.contexts)-1]
	}
	return context{}
}

func (t *contextStack) IsEmpty() bool {
	if len(t.contexts) > 0 {
		return false
	}
	return true
}
