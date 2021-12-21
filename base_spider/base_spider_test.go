package base_spider

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"testing"

	"github.com/obgnail/ScrapyInGo/entity"
)

var favoritesBookRegexp = regexp.MustCompile(`<a href="/g/(\d+?)/" class="cover".*?>`)
var lastBookPageRegexp = regexp.MustCompile(`<a href="/favorites/?page=(\d+?)" class="last">`)

type NhentaiSpider struct {
	cookieString string
	*BaseSpider
}

func NewMySpider(name string, threadNum uint, cookieString string) *NhentaiSpider {
	return &NhentaiSpider{cookieString: cookieString, BaseSpider: NewBaseSpider(name, threadNum)}
}

func (s *NhentaiSpider) getFavoritesPageRequest() (*entity.Request, error) {
	reqObj, err := http.NewRequest("GET", "https://nhentai.net/favorites/", nil)
	if err != nil {
		return nil, err
	}
	reqObj.Header.Set("cookie", s.cookieString)
	reqObj.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.110 Safari/537.36")

	req := entity.NewRequest(reqObj, 100, false, s.Parse, s.ErrBack, nil)
	return req, nil
}

func (s *NhentaiSpider) Parse(resp *entity.Response) (interface{}, error) {
	if resp.GetStatus() != 200 {
		log.Fatalln(resp.GetContent())
	}
	content, err := resp.GetContent()
	if err != nil {
		return nil, err
	}
	page := favoritesBookRegexp.FindAllSubmatch(content,1)
	fmt.Sprintln(page)
	return nil,nil
}

func (s *NhentaiSpider) ErrBack(r *entity.Response, err error) {
	return
}

func (s *NhentaiSpider) ParseIndexPage(res *entity.Response) error {
	return nil
}

func (s *NhentaiSpider) ParseDetailPage(res *entity.Response) {
	return
}

func (s *NhentaiSpider) Errback(res *entity.Response, err error) {
	return
}

func TestSpider(t *testing.T) {
	cookieStr := "cf_clearance=daa9d77091a0e1b953eaeff3cf032ad0d014e6f4-1626532362-0-150; csrftoken=Zxvfpgcdhf1uDMzzlwvsB4gC7TICWi6TRZcCubCnfEnECwY8MzfVbXgg6Bzw2Ytf; sessionid=uzckbjtbrie0smci1z0z5y13de5nf59g"
	spider := NewMySpider("nhentaiSpider", 3, cookieStr)
	req, err := spider.getFavoritesPageRequest()
	if err != nil {
		log.Fatalln(err)
	}
	spider.PushRequest(req)
	spider.Crawl()
}
