package demo

import (
	"bytes"
	"fmt"
	"github.com/juju/errors"
	"github.com/obgnail/ScrapyInGo/core/downloader"
	"github.com/obgnail/ScrapyInGo/core/entity"
	"github.com/obgnail/ScrapyInGo/core/middleware"
	"github.com/obgnail/ScrapyInGo/core/spider_imp"
	"log"
	"path/filepath"
	"regexp"
	"strconv"
	"sync"
	"time"
)

var (
	favoritesBookRegexp  = regexp.MustCompile(`<a href="/g/(\d+?)/" class="cover".*?>`)
	lastBookPageRegexp   = regexp.MustCompile(`<a href="/favorites/\?page=(\d+?)" class="last">`)
	mangaPageNumRegexp   = regexp.MustCompile(`Pages:(?s:.*?)<span class="name">(\d+?)</span>`)
	mangaUrlSuffixRegexp = regexp.MustCompile(`<a class="gallerythumb".*?data-src="https://t\.nhentai\.net/galleries/(\d+?)/.+?\.(.+?)".*?/>`)
	mangaNameRegexp      = regexp.MustCompile(`<span class="before">(.*?)</span><span class="pretty">(.*?)</span><span class="after">(.*?)</span>`)
	ValidNameRegexp      = regexp.MustCompile(`[\\/:*?"<>|\r\n]+`)
)

var ParseOnce sync.Once

type Manga struct {
	content  []byte
	fileName string
	dirName  string
}

type NhentaiSpider struct {
	*spider_imp.SimpleSpider
	storeBasePath string
}

func New(name string, threadNum uint, headers map[string]string, proxy string, storeBasePath string) *NhentaiSpider {
	sp := &NhentaiSpider{SimpleSpider: spider_imp.NewSimpleSpider(name, threadNum)}
	sp.SetDownloader(downloader.NewProxyDownloader(proxy))
	sp.SetPipeline(NewStoreMangaPipeline(storeBasePath))
	sp.AppendMiddleware(middleware.NewSetHeaderMiddleware(headers))
	sp.storeBasePath = storeBasePath
	return sp
}

func (s *NhentaiSpider) QueueStartUrl() {
	url := "https://nhentai.net/favorites/"
	req, err := newSimpleRequest(url, s.Parse, s.ErrBack, nil)
	if err != nil {
		log.Fatalln(errors.Trace(err))
	}
	s.QueueRequest(req)
}

func (s *NhentaiSpider) Parse(resp *entity.Response) (interface{}, error) {
	content, err := getContentFromResponse(resp)
	if err != nil {
		return nil, errors.Trace(err)
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
			return nil, errors.Trace(e)
		}
		s.QueueRequest(mangaIndexPageReq)
	}

	ParseOnce.Do(func() {
		favoritesPageMatch := lastBookPageRegexp.FindSubmatch(content)
		if favoritesPageMatch == nil {
			return
		}
		lastPage := string(favoritesPageMatch[1])
		if lastPage == "" {
			log.Printf("[WARN] get last page error,url:[%s]", resp.Request.GetUrl())
			return
		}
		lp, err := strconv.Atoi(lastPage)
		if err != nil {
			log.Printf("[WARN] get last page error,url:[%s]", resp.Request.GetUrl())
			return
		}
		for i := 2; i <= lp; i++ {
			url := fmt.Sprintf("https://nhentai.net/favorites/?page=%d", i)
			favoritesPageReq, e := newSimpleRequest(url, s.Parse, s.ErrBack, nil)
			if e != nil {
				return
			}
			s.QueueRequest(favoritesPageReq)
		}
	})

	return nil, nil
}

func (s *NhentaiSpider) ParseMangaIndexPage(resp *entity.Response) (interface{}, error) {
	content, err := getContentFromResponse(resp)
	if err != nil {
		return nil, errors.Trace(err)
	}

	mangaNameMatch := mangaNameRegexp.FindSubmatch(content)
	name := bytes.Join(mangaNameMatch[1:], []byte(""))
	mangeDirName := ValidName(name)

	// 如果目标目录已经存在,则不下载此漫画
	dirPath := filepath.Join(s.storeBasePath, string(mangeDirName))
	if IsDirExist(dirPath) {
		return nil, nil
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

	// 1t.jpg
	mangaUrlSuffixMatch := mangaUrlSuffixRegexp.FindSubmatch(content)
	mangaID, suffix := string(mangaUrlSuffixMatch[1]), string(mangaUrlSuffixMatch[2])
	if suffix == "" || mangaID == "" {
		log.Printf("[WARN] get manga suffix error,url:[%s]", resp.Request.GetUrl())
		return nil, nil
	}

	for i := 1; i <= pages; i++ {
		fileName := fmt.Sprintf("%d.%s", i, suffix)
		url := fmt.Sprintf("https://i.nhentai.net/galleries/%s/%s", mangaID, fileName)
		meta := map[string]interface{}{"mangaDirName": string(mangeDirName), "mangaFileName": fileName}
		mangaReq, e := newSimpleRequest(url, s.ParseMangaContent, s.ErrBack, meta)
		if e != nil {
			return nil, errors.Trace(e)
		}
		s.QueueRequest(mangaReq)
	}

	return nil, nil
}

// ErrBack 请求错误、解析错误都会回调此函数
func (s *NhentaiSpider) ErrBack(req *entity.Request, err error) {
	req.IncReTries()
	url := req.GetReqObj().URL
	retries := req.GetReTries()

	log.Printf("[ERROR] error back, retries: %d , url: %s\n", retries, url)
	log.Println("[ERROR] Trace: ->", errors.Trace(err))

	if retries > 5 {
		log.Printf("reTry too much: -> %s\n", url)
		return
	}
	time.Sleep(10 * time.Second)
	log.Printf("requeue url: %s\n", url)
	s.QueueRequest(req)
}

func (s *NhentaiSpider) ParseMangaContent(resp *entity.Response) (interface{}, error) {
	content, err := getContentFromResponse(resp)
	if err != nil {
		return nil, errors.Trace(err)
	}
	manga := &Manga{
		content:  content,
		dirName:  resp.Meta["mangaDirName"].(string),
		fileName: resp.Meta["mangaFileName"].(string),
	}
	return manga, nil
}

func (s *NhentaiSpider) Run() {
	s.QueueStartUrl()
	s.Crawl()
}
