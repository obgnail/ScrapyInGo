package scheduler

import (
	"github.com/obgnail/ScrapyInGo/core/entity"
)

type Scheduler interface {
	Push(*entity.Request)
	Pop() *entity.Request
	Len() int
}
