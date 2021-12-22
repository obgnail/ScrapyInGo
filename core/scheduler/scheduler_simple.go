package scheduler

import (
	"github.com/obgnail/ScrapyInGo/core/entity"
)

const (
	chanSize = 1024
)

type SimpleScheduler struct {
	queue chan *entity.Request
}

func NewSimpleScheduler() *SimpleScheduler {
	return &SimpleScheduler{make(chan *entity.Request, chanSize)}
}

func (s *SimpleScheduler) Push(req *entity.Request) {
	s.queue <- req
}

func (s *SimpleScheduler) Pop() *entity.Request {
	return <-s.queue
}

func (s *SimpleScheduler) Len() int {
	return len(s.queue)
}
