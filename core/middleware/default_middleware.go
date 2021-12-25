package middleware

import (
	"github.com/obgnail/ScrapyInGo/core/entity"
	"github.com/obgnail/ScrapyInGo/core/spider"
)

type DefaultMiddleware struct{}

func (m *DefaultMiddleware) ProcessRequest(req *entity.Request, sp spider.Spider) *entity.Request {
	return req
}
func (m *DefaultMiddleware) ProcessResponse(resp *entity.Response, sp spider.Spider) *entity.Response {
	return resp
}
func (m *DefaultMiddleware) ProcessItem(item interface{}, sp spider.Spider) interface{} {
	return item
}
func (m *DefaultMiddleware) SpiderOpened(sp spider.Spider) {}

func NewDefaultMiddleware() *DefaultMiddleware {
	return &DefaultMiddleware{}
}
