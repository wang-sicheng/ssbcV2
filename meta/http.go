package meta

type HttpResponse struct {
	StatusCode int
	Data       interface{}
	Code 	   int	`json:"code"`	// 前端code校验
}
