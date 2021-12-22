package engine

import (
	"log"
	"time"

	"github.com/juju/errors"

	"github.com/obgnail/ScrapyInGo/core/downloader"
	"github.com/obgnail/ScrapyInGo/core/entity"
	"github.com/obgnail/ScrapyInGo/core/middleware"
	"github.com/obgnail/ScrapyInGo/core/pipeline"
	"github.com/obgnail/ScrapyInGo/core/scheduler"
	"github.com/obgnail/ScrapyInGo/core/spider"
)

const (
	ReqChanMaxLen  = 4
	RespChanMaxLen = 100
	ItemChanMaxLen = 300

	SleepTimeWhenHasNoRequest = 100 * time.Millisecond
)

type Engine struct {
	spider      spider.Spider
	scheduler   scheduler.Scheduler
	downloader  downloader.Downloader
	pipeline    pipeline.Pipeline
	middlewares []middleware.Middleware

	reqChan   chan *entity.Request
	respChan  chan *entity.Response
	itemChan  chan interface{}
	closeChan chan struct{}
}

func New(
	sp spider.Spider,
	scheduler scheduler.Scheduler,
	downloader downloader.Downloader,
	pipeline pipeline.Pipeline,
	middleware []middleware.Middleware,
) *Engine {
	e := &Engine{
		spider:      sp,
		scheduler:   scheduler,
		downloader:  downloader,
		pipeline:    pipeline,
		middlewares: middleware,
		reqChan:     make(chan *entity.Request, ReqChanMaxLen),
		respChan:    make(chan *entity.Response, RespChanMaxLen),
		itemChan:    make(chan interface{}, ItemChanMaxLen),
		closeChan:   make(chan struct{}, 1),
	}
	return e
}

func Default(sp spider.Spider) *Engine {
	return New(
		sp,
		scheduler.NewSimpleScheduler(),
		downloader.NewDefaultDownloader(),
		pipeline.NewDefaultPipeline(),
		[]middleware.Middleware{middleware.NewDefaultMiddleware()},
	)
}

func (e *Engine) PushRequest(req *entity.Request) {
	if e.scheduler == nil {
		log.Fatal("spider has no scheduler")
	}
	e.scheduler.Push(req)
}

func (e *Engine) PushRequests(requests []*entity.Request) {
	if e.scheduler == nil {
		log.Fatal("spider has no scheduler")
	}
	for _, req := range requests {
		e.scheduler.Push(req)
	}
}

func (e *Engine) SetThreadLimit(num uint) {
	if len(e.reqChan) != 0 {
		log.Fatal("req chan is not empty")
	}
	e.reqChan = make(chan *entity.Request, num)
}

func (e *Engine) SetScheduler(scheduler scheduler.Scheduler) {
	e.scheduler = scheduler
}

func (e *Engine) SetDownloader(downloader downloader.Downloader) {
	e.downloader = downloader
}

func (e *Engine) SetPipeline(pipeline pipeline.Pipeline) {
	e.pipeline = pipeline
}

func (e *Engine) AppendMiddleware(middleware middleware.Middleware) {
	e.middlewares = append(e.middlewares, middleware)
}

func (e *Engine) GetScheduler() scheduler.Scheduler {
	return e.scheduler
}

func (e *Engine) GetDownloader() downloader.Downloader {
	return e.downloader
}

func (e *Engine) GetPipeline() pipeline.Pipeline {
	return e.pipeline
}

func (e *Engine) GetMiddleware() []middleware.Middleware {
	return e.middlewares
}

func (e *Engine) startScheduler() {
	if e.scheduler == nil {
		log.Fatal("engine has no scheduler")
	}
	for {
		if e.scheduler.Len() != 0 {
			e.reqChan <- e.scheduler.Pop()
		} else {
			time.Sleep(SleepTimeWhenHasNoRequest)
		}
	}
}

func (e *Engine) startDownloader() {
	if e.downloader == nil {
		log.Fatal("engine has no downloader")
	}
	for {
		select {
		case req := <-e.reqChan:
			go func() {
				for _, m := range e.middlewares {
					req = m.ProcessRequest(req, e.spider)
				}
				resp, err := e.downloader.Download(req)
				if err != nil {
					log.Println("[WARN] download error", errors.Trace(err))
					resp.GetRespObj().Body.Close()
					if req.Errback != nil {
						req.Errback(req, err)
					}
					return
				}
				for _, m := range e.middlewares {
					resp = m.ProcessResponse(resp, e.spider)
				}
				if resp != nil {
					e.respChan <- resp
				}
			}()
		}
	}
}

func (e *Engine) startParse() {
	if e.spider == nil {
		log.Fatal("engine has no spider")
	}
	var ParseFunc func(r *entity.Response) (interface{}, error)
	for {
		select {
		case resp := <-e.respChan:
			go func() {
				if resp.Callback == nil {
					ParseFunc = e.spider.Parse
				} else {
					ParseFunc = resp.Callback
				}
				item, err := ParseFunc(resp)
				resp.GetRespObj().Body.Close()
				if err != nil {
					log.Println("[WARN] parse error", errors.Trace(err))
					if resp.Errback != nil {
						resp.Errback(resp.Request, err)
					}
					return
				}
				if item != nil {
					e.itemChan <- item
				}
			}()
		}
	}
}

func (e *Engine) startPipeline() {
	if e.pipeline == nil {
		log.Fatal("engine has no pipeline")
	}
	for {
		select {
		case item := <-e.itemChan:
			go func() {
				if err := e.pipeline.ProcessItem(item, e.spider); err != nil {
					log.Printf("processItem err %s", err)
				}
			}()
		}
	}
}

func (e *Engine) Crawl() {
	for _, m := range e.middlewares {
		m.SpiderOpened(e.spider)
	}
	go e.startPipeline()
	go e.startParse()
	go e.startDownloader()
	go e.startScheduler()
	<-e.closeChan
}
