package downloader

import (
	"net/http"
	"net/url"

	"github.com/juju/errors"

	"github.com/obgnail/ScrapyInGo/core/entity"
)

type ProxyDownloader struct {
	ProxyServer string
}

func NewProxyDownloader(proxyServer string) *ProxyDownloader {
	return &ProxyDownloader{proxyServer}
}

func (d *ProxyDownloader) ProxyClient() *http.Client {
	proxyURL, _ := url.Parse(d.ProxyServer)
	return &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyURL)}}
}

func (d *ProxyDownloader) Download(req *entity.Request) (*entity.Response, error) {
	client := d.ProxyClient()
	resp, err := entity.ProcessRequest(client, req)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return resp, nil
}
