package middleware

import (
	"github.com/obgnail/ScrapyInGo/core/entity"
	"github.com/obgnail/ScrapyInGo/core/spider"
)

type SetHeaderMiddleware struct {
	headers map[string]string
}

func (m *SetHeaderMiddleware) ProcessRequest(req *entity.Request, sp spider.Spider) *entity.Request {
	for key, val := range m.headers {
		req.GetReqObj().Header.Set(key, val)
	}
	return req
}
func (m *SetHeaderMiddleware) ProcessResponse(resp *entity.Response, sp spider.Spider) *entity.Response {
	return resp
}
func (m *SetHeaderMiddleware) SpiderOpened(sp spider.Spider) {}

func NewSetHeaderMiddleware(headers map[string]string) *SetHeaderMiddleware {
	return &SetHeaderMiddleware{headers: headers}
}
