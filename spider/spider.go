package spider

import (
	"github.com/obgnail/ScrapyInGo/entity"
)

type Spider interface {
	Parse(*entity.Response) (interface{}, error)
}
