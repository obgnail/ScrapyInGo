package scheduler

import (
	"container/list"
	"crypto/md5"
	"github.com/obgnail/ScrapyInGo/core/entity"
	"sync"
)

type QueueScheduler struct {
	locker *sync.Mutex
	rm     bool
	rmKey  map[[md5.Size]byte]*list.Element
	queue  *list.List
}

func NewQueueScheduler(rmDuplicate bool) *QueueScheduler {
	queue := list.New()
	rmKey := make(map[[md5.Size]byte]*list.Element)
	locker := new(sync.Mutex)
	return &QueueScheduler{rm: rmDuplicate, queue: queue, rmKey: rmKey, locker: locker}
}

func (s *QueueScheduler) Push(req *entity.Request) {
	s.locker.Lock()
	var key [md5.Size]byte
	if s.rm {
		key = md5.Sum([]byte(req.GetUrl()))
		if _, ok := s.rmKey[key]; ok {
			s.locker.Unlock()
			return
		}
	}
	e := s.queue.PushBack(req)
	if s.rm {
		s.rmKey[key] = e
	}
	s.locker.Unlock()
}

func (s *QueueScheduler) Pop() *entity.Request {
	s.locker.Lock()
	if s.queue.Len() <= 0 {
		s.locker.Unlock()
		return nil
	}
	e := s.queue.Front()
	req := e.Value.(*entity.Request)
	key := md5.Sum([]byte(req.GetUrl()))
	s.queue.Remove(e)
	if s.rm {
		delete(s.rmKey, key)
	}
	s.locker.Unlock()
	return req
}

func (s *QueueScheduler) Len() int {
	s.locker.Lock()
	length := s.queue.Len()
	s.locker.Unlock()
	return length
}
