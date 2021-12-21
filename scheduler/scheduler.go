package scheduler

import (
	"github.com/obgnail/ScrapyInGo/entity"
)

type Scheduler interface {
	Push(*entity.Request)
	Pop() *entity.Request
	Len() int
}
