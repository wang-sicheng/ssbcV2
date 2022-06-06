package meta

// TCP传递消息
type TCPMessage struct {
	Type    string //消息类型
	Content []byte //消息体
	From    string
	To      string
}

type HttpResponse struct {
	Error string      `json:"error"` // 如果不为空代表错误信息
	Data  interface{} `json:"data"`
	Code  int         `json:"code"` // vue-element-admin的前端校验码，必须为20000
}
