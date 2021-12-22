package demo

import (
	"bytes"
	"fmt"
	"github.com/obgnail/ScrapyInGo/core/downloader"
	"github.com/obgnail/ScrapyInGo/core/entity"
	"github.com/obgnail/ScrapyInGo/core/middleware"
	"github.com/obgnail/ScrapyInGo/core/spider_imp"
	"log"
	"regexp"
	"strconv"
)

const (
	UA     = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.110 Safari/537.36"
	Cookie = "cf_clearance=daa9d77091a0e1b953eaeff3cf032ad0d014e6f4-1626532362-0-150; csrftoken=Zxvfpgcdhf1uDMzzlwvsB4gC7TICWi6TRZcCubCnfEnECwY8MzfVbXgg6Bzw2Ytf; sessionid=uzckbjtbrie0smci1z0z5y13de5nf59g"
)

var (
	favoritesBookRegexp  = regexp.MustCompile(`<a href="/g/(\d+?)/" class="cover".*?>`)
	lastBookPageRegexp   = regexp.MustCompile(`<a href="/favorites/\?page=(\d+?)" class="last">`)
	mangaPageNumRegexp   = regexp.MustCompile(`Pages:(?s:.*?)<span class="name">(\d+?)</span>`)
	mangaUrlSuffixRegexp = regexp.MustCompile(`<a class="gallerythumb".*?data-src="https://t\.nhentai\.net/galleries/(\d+?)/.+?\.(.+?)".*?/>`)
	mangaNameRegexp      = regexp.MustCompile(`<span class="before">(.*?)</span><span class="pretty">(.*?)</span><span class="after">(.*?)</span>`)
	ValidNameRegexp      = regexp.MustCompile(`[\\/:*?"<>|\r\n]+`)
)

type Manga struct {
	content  []byte
	fileName string
	dirName  string
}

type NhentaiSpider struct {
	*spider_imp.SimpleSpider
}

func New(name string, threadNum uint, headers map[string]string, proxy string) *NhentaiSpider {
	sp := &NhentaiSpider{SimpleSpider: spider_imp.NewSimpleSpider(name, threadNum)}
	sp.SetDownloader(downloader.NewProxyDownloader(proxy))
	sp.SetPipeline(NewStoreMangaPipeline())
	sp.AppendMiddleware(middleware.NewSetHeaderMiddleware(headers))
	return sp
}

func (s *NhentaiSpider) getFavoritesPageRequest() (*entity.Request, error) {
	url := "https://nhentai.net/favorites/"
	return newSimpleRequest(url, s.Parse, s.ErrBack, nil)
}

func (s *NhentaiSpider) Parse(resp *entity.Response) (interface{}, error) {
	content, err := getContentFromResponse(resp)
	if err != nil {
		return nil, err
	}
	mangaMatches := favoritesBookRegexp.FindAllSubmatch(content, -1)
	for _, match := range mangaMatches {
		mangaID := string(match[1])
		if mangaID == "" {
			log.Printf("[WARN] get favorites book error,url:[%s]", resp.Request.GetUrl())
			continue
		}
		url := fmt.Sprintf("https://nhentai.net/g/%s/", mangaID)
		mangaIndexPageReq, e := newSimpleRequest(url, s.ParseMangaIndexPage, s.ErrBack, nil)
		if e != nil {
			return nil, e
		}
		s.PushRequest(mangaIndexPageReq)
	}

	favoritesPageMatch := lastBookPageRegexp.FindSubmatch(content)
	lastPage := string(favoritesPageMatch[1])
	if lastPage == "" {
		log.Printf("[WARN] get last page error,url:[%s]", resp.Request.GetUrl())
		return nil, nil
	}
	lp, err := strconv.Atoi(lastPage)
	if err != nil {
		log.Printf("[WARN] get last page error,url:[%s]", resp.Request.GetUrl())
		return nil, nil
	}
	for i := 2; i <= lp; i++ {
		url := fmt.Sprintf("https://nhentai.net/favorites/?page=%d", i)
		favoritesPageReq, e := newSimpleRequest(url, s.Parse, s.ErrBack, nil)
		if e != nil {
			return nil, e
		}
		s.PushRequest(favoritesPageReq)
	}
	return nil, nil
}

func (s *NhentaiSpider) ParseMangaIndexPage(resp *entity.Response) (interface{}, error) {
	content, err := getContentFromResponse(resp)
	if err != nil {
		return nil, err
	}

	pageNumMatch := mangaPageNumRegexp.FindSubmatch(content)
	pageNum := string(pageNumMatch[1])
	if pageNum == "" {
		log.Printf("[WARN] get page num error,url:[%s]", resp.Request.GetUrl())
		return nil, nil
	}
	pages, err := strconv.Atoi(pageNum)
	if err != nil {
		log.Printf("[WARN] get last page error,url:[%s]", resp.Request.GetUrl())
		return nil, nil
	}

	mangaNameMatch := mangaNameRegexp.FindSubmatch(content)
	name := bytes.Join(mangaNameMatch[1:], []byte(""))
	mangeDirName := ValidName(name)
	meta := map[string]interface{}{"mangaDirName": mangeDirName, "mangaFileName": ""}

	// 1t.jpg
	mangaUrlSuffixMatch := mangaUrlSuffixRegexp.FindSubmatch(content)
	mangaID, suffix := string(mangaUrlSuffixMatch[1]), string(mangaUrlSuffixMatch[2])
	if suffix == "" || mangaID == "" {
		log.Printf("[WARN] get manga suffix error,url:[%s]", resp.Request.GetUrl())
		return nil, nil
	}

	for i := 1; i <= pages; i++ {
		url := fmt.Sprintf("https://i.nhentai.net/galleries/%s/%d.%s", mangaID, i, suffix)
		meta["mangaFileName"] = fmt.Sprintf("%d", i)
		mangaReq, e := newSimpleRequest(url, s.ParseMangaContent, s.ErrBack, meta)
		if e != nil {
			return nil, e
		}
		s.PushRequest(mangaReq)
	}

	return nil, nil
}

func (s *NhentaiSpider) ErrBack(req *entity.Request, err error) {
	fmt.Println("errback ->", err)
	return
}

func (s *NhentaiSpider) ParseMangaContent(resp *entity.Response) (interface{}, error) {
	content, err := getContentFromResponse(resp)
	if err != nil {
		return nil, err
	}
	manga := &Manga{
		content:  content,
		dirName:  resp.Meta["mangaDirName"].(string),
		fileName: resp.Meta["mangaFileName"].(string),
	}
	return manga, nil
}

func (s *NhentaiSpider) Run() {
	req, err := s.getFavoritesPageRequest()
	if err != nil {
		log.Fatalln(err)
	}
	s.PushRequest(req)
	s.Crawl()
}

func Run() {
	spider := New(
		"nhentai",
		1,
		map[string]string{"cookie": Cookie, "user-agent": UA},
		"http://127.0.0.1:7890",
	)
	spider.Run()
}
