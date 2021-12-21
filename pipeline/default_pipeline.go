package pipeline

import "fmt"

type DefaultPipeline struct{}

func (p *DefaultPipeline) ProcessItem(item interface{}) error {
	fmt.Println(item)
	return nil
}

func NewDefaultPipeline() *DefaultPipeline {
	return &DefaultPipeline{}
}
