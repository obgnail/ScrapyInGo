package downloader

import (
	"net/http"

	"github.com/juju/errors"

	"github.com/obgnail/ScrapyInGo/entity"
)

type DefaultDownloader struct{}

func NewDefaultDownloader() *DefaultDownloader {
	return &DefaultDownloader{}
}
func (d *DefaultDownloader) Download(req *entity.Request) (*entity.Response, error) {
	client := &http.Client{}
	respObj, err := client.Do(req.GetReqObj())
	if err != nil {
		return nil, errors.Errorf("download failed, err:[%s]", err.Error())
	}
	defer respObj.Body.Close()

	resp := entity.FromRequest(req)
	resp.SetRespObj(respObj)
	return resp, nil
}
