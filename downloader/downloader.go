package downloader

import (
	"github.com/obgnail/ScrapyInGo/entity"
)

type Downloader interface {
	Download(req *entity.Request) (*entity.Response, error)
}
