package base_spider

import (
	"fmt"
	"github.com/obgnail/ScrapyInGo/engine"
	"github.com/obgnail/ScrapyInGo/entity"
)

type BaseSpider struct {
	Name       string
	TheadLimit uint

	*engine.Engine
}

func NewBaseSpider(name string, theadLimit uint) *BaseSpider {
	s := &BaseSpider{Name: name}
	s.Engine = engine.Default(s)
	s.SetThreadLimit(theadLimit)
	return s
}

func (s *BaseSpider) Parse(resp *entity.Response) (interface{}, error) {
	fmt.Print(resp)
	return nil, nil
}
