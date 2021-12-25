package downloader

import (
	"github.com/juju/errors"
	"net/http"

	"github.com/obgnail/ScrapyInGo/core/entity"
)

type DefaultDownloader struct{}

func NewDefaultDownloader() *DefaultDownloader {
	return &DefaultDownloader{}
}
func (d *DefaultDownloader) Download(req *entity.Request) (*entity.Response, error) {
	client := &http.Client{}
	resp, err := entity.ProcessRequest(client, req)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return resp, nil
}
