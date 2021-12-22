package pipeline

import (
	"github.com/obgnail/ScrapyInGo/core/spider"
)

type Pipeline interface {
	ProcessItem(item interface{}, sp spider.Spider) error
}
