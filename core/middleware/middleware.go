package middleware

import (
	"github.com/obgnail/ScrapyInGo/core/entity"
	"github.com/obgnail/ScrapyInGo/core/spider"
)

type Middleware interface {
	ProcessRequest(req *entity.Request, sp spider.Spider) *entity.Request
	ProcessResponse(resp *entity.Response, sp spider.Spider) *entity.Response
	ProcessItem(item interface{}, sp spider.Spider) interface{}
	SpiderOpened(sp spider.Spider)
}
