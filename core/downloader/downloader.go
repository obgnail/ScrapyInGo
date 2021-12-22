package downloader

import (
	"github.com/obgnail/ScrapyInGo/core/entity"
)

type Downloader interface {
	Download(req *entity.Request) (*entity.Response, error)
}
