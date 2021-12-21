package middleware

import (
	"github.com/obgnail/ScrapyInGo/entity"
	"github.com/obgnail/ScrapyInGo/spider"
)

type Middleware interface {
	ProcessRequest(req *entity.Request, sp spider.Spider) *entity.Request
	ProcessResponse(resp *entity.Response, sp spider.Spider) *entity.Response
	SpiderOpened(sp spider.Spider)
}
