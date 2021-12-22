package spider_imp

import (
	"github.com/obgnail/ScrapyInGo/core/engine"
	"github.com/obgnail/ScrapyInGo/core/entity"
	"log"
)

type SimpleSpider struct {
	Name       string
	TheadLimit uint

	*engine.Engine
}

func NewSimpleSpider(name string, theadLimit uint) *SimpleSpider {
	s := &SimpleSpider{Name: name}
	s.Engine = engine.Default(s)
	s.SetThreadLimit(theadLimit)
	return s
}

func (s *SimpleSpider) Parse(resp *entity.Response) (interface{}, error) {
	log.Fatalln("has not implement spider interface")
	return nil, nil
}
