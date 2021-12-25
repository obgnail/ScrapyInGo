package utils

import "sync"

type Semaphore struct {
	blockChan chan struct{}
	wg        *sync.WaitGroup
}

func NewSemaphore(maxSize int) *Semaphore {
	return &Semaphore{
		blockChan: make(chan struct{}, maxSize),
		wg:        new(sync.WaitGroup),
	}
}

func (s *Semaphore) Add(delta int) {
	s.wg.Add(delta)
	for i := 0; i < delta; i++ {
		s.blockChan <- struct{}{}
	}
}

func (s *Semaphore) Done() {
	<-s.blockChan
	s.wg.Done()
}

func (s *Semaphore) Wait() {
	s.wg.Wait()
}
