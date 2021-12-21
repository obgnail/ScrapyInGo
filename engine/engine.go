package engine

import (
	"log"
	"time"

	"github.com/juju/errors"

	"github.com/obgnail/ScrapyInGo/downloader"
	"github.com/obgnail/ScrapyInGo/entity"
	"github.com/obgnail/ScrapyInGo/middleware"
	"github.com/obgnail/ScrapyInGo/pipeline"
	"github.com/obgnail/ScrapyInGo/scheduler"
	"github.com/obgnail/ScrapyInGo/spider"
)

const (
	ReqChanMaxLen  = 4
	RespChanMaxLen = 100
	ItemChanMaxLen = 300

	SleepTimeWhenHasNoRequest = 100 * time.Millisecond
)

type Engine struct {
	Spider      spider.Spider
	Scheduler   scheduler.Scheduler
	Downloader  downloader.Downloader
	Pipeline    pipeline.Pipeline
	Middlewares []middleware.Middleware

	ReqChan  chan *entity.Request
	RespChan chan *entity.Response
	ItemChan chan interface{}
}

func NewEngine(
	sp spider.Spider,
	scheduler scheduler.Scheduler,
	downloader downloader.Downloader,
	pipeline pipeline.Pipeline,
	middleware []middleware.Middleware,
) *Engine {
	e := &Engine{
		Spider:      sp,
		Scheduler:   scheduler,
		Downloader:  downloader,
		Pipeline:    pipeline,
		Middlewares: middleware,
		ReqChan:     make(chan *entity.Request, ReqChanMaxLen),
		RespChan:    make(chan *entity.Response, RespChanMaxLen),
		ItemChan:    make(chan interface{}, ItemChanMaxLen),
	}
	return e
}

func Default(sp spider.Spider) *Engine {
	return NewEngine(
		sp,
		scheduler.NewSimpleScheduler(),
		downloader.NewDefaultDownloader(),
		pipeline.NewDefaultPipeline(),
		[]middleware.Middleware{middleware.NewDefaultMiddleware()},
	)
}

func (e *Engine) PushRequest(req *entity.Request) {
	if e.Scheduler == nil {
		log.Fatal("spider has no scheduler")
	}
	e.Scheduler.Push(req)
}

func (e *Engine) PushRequests(requests []*entity.Request) {
	if e.Scheduler == nil {
		log.Fatal("spider has no scheduler")
	}
	for _, req := range requests {
		e.Scheduler.Push(req)
	}
}

func (e *Engine) SetThreadLimit(num uint) {
	if len(e.ReqChan) != 0 {
		log.Fatal("req chan is not empty")
	}
	e.ReqChan = make(chan *entity.Request, num)
}

func (e *Engine) SetScheduler(scheduler scheduler.Scheduler) {
	e.Scheduler = scheduler
}

func (e *Engine) SetDownloader(downloader downloader.Downloader) {
	e.Downloader = downloader
}

func (e *Engine) AppendMiddleware(middleware middleware.Middleware) {
	e.Middlewares = append(e.Middlewares, middleware)
}

func (e *Engine) startScheduler() {
	if e.Scheduler == nil {
		log.Fatal("engine has no scheduler")
	}
	for {
		if e.Scheduler.Len() != 0 {
			e.ReqChan <- e.Scheduler.Pop()
		} else {
			time.Sleep(SleepTimeWhenHasNoRequest)
		}
	}
}

func (e *Engine) startDownloader() {
	if e.Downloader == nil {
		log.Fatal("engine has no downloader")
	}
	for {
		select {
		case req := <-e.ReqChan:
			go func() {
				for _, m := range e.Middlewares {
					req = m.ProcessRequest(req, e.Spider)
				}
				resp, err := e.Downloader.Download(req)
				if err != nil {
					log.Println("[WARN] download error", errors.Trace(err))
					if req.Errback != nil {
						req.Errback(resp, err)
					}
					return
				}
				for _, m := range e.Middlewares {
					resp = m.ProcessResponse(resp, e.Spider)
				}
				e.RespChan <- resp
			}()
		}
	}
}

func (e *Engine) startParse() {
	if e.Spider == nil {
		log.Fatal("engine has no Spider")
	}
	var ParseFunc func(r *entity.Response) (interface{}, error)
	for {
		select {
		case resp := <-e.RespChan:
			go func() {
				if resp.Callback == nil {
					ParseFunc = e.Spider.Parse
				} else {
					ParseFunc = resp.Callback
				}
				item, err := ParseFunc(resp)
				if err != nil {
					log.Println("[WARN] parse error", errors.Trace(err))
					if resp.Errback != nil {
						resp.Errback(resp, err)
					}
				}
				e.ItemChan <- item
			}()
		}
	}
}

func (e *Engine) startPipeline() {
	if e.Pipeline == nil {
		log.Fatal("engine has no pipeline")
	}
	for {
		select {
		case item := <-e.ItemChan:
			go func() {
				if err := e.Pipeline.ProcessItem(item); err != nil {
					log.Printf("processItem err %s", err)
				}
			}()
		}
	}
}

func (e *Engine) Crawl() {
	for _, m := range e.Middlewares {
		m.SpiderOpened(e.Spider)
	}
	go e.startPipeline()
	go e.startParse()
	go e.startDownloader()
	go e.startScheduler()
	select {}
}
