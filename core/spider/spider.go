package spider

import (
	"github.com/obgnail/ScrapyInGo/core/entity"
)

type Spider interface {
	Parse(*entity.Response) (interface{}, error)
}
