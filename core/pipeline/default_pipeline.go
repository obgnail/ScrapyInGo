package pipeline

import (
	"fmt"
	"github.com/obgnail/ScrapyInGo/core/spider"
)

type DefaultPipeline struct{}

func (p *DefaultPipeline) ProcessItem(item interface{}, sp spider.Spider) error {
	fmt.Println("pipeline ->", item)
	return nil
}

func NewDefaultPipeline() *DefaultPipeline {
	return &DefaultPipeline{}
}
