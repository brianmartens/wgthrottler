package wgthrottler

import (
	"sync"
)

// WgThrottler - A throttled waitgroup for limiting concurrent/parallel processes.
//  wg - Underlying sync.WaitGroup used for tracking goroutines
//  c - Active count of running processes
//  max - Maximum allowed number of active processes
//  ch - Channel used to communicate when a process is complete
//  mu - mutex to block multiple goroutines from receiving over ch
type WgThrottler struct {
	wg  sync.WaitGroup
	c   int
	max int
	ch  chan int
	mu  sync.Mutex
}

// Throttler is an interface which expects three methods: Done(), Wait(), and Next().
// Done() and Wait() should function equivalently to a sync.WaitGroup, whereas Next() blocks until a new goroutine
// may be allocated according to an arbitrary ruleset defined by the implementation.
type Throttler interface{
	Done()
	Wait()
	Next()
}

// Done is functionally equivalent to a sync.WaitGroup's Done() method.
// A value of -1 will be sent through ch and the underlying sync.WaitGroup
func (wg *WgThrottler) Done() {
	// Defer a function closure to pass -1 to ch and then call wg.wg.Done().
	// .Done() must be called first because we call .Wait() first in the
	// WhThrottler's Wait() method below.
	defer func() {
		wg.ch <- -1
	}()
	wg.wg.Done()
}

// Wait is functionally equivalent to a regular sync.WaitGroup's Wait() method.
// This will force the WgThrottler to wait until all running goroutines have completed.
func (wg *WgThrottler) Wait() {
	// Wait for all goroutines to call wg.Done()
	wg.wg.Wait()
	// Collect any lingering channel values waiting to be received.
	for wg.c > 0 {
		wg.c += <-wg.ch
	}
}

// Next will iterate the WgThrottler to the next allowed process.
//
//  for i := 0; i < 10; i++ {
//    go func(){
//        defer wg.Next()
// 		  MyFunc()
//    }
//  }
func (wg *WgThrottler) Next() {
	// Lock the mutex to prevent concurrent reads and writes to our counter c
	wg.mu.Lock()
	// Unlock the mutex upon exit.
	defer wg.mu.Unlock()
	// If we have reached our concurrency limit, then accept values over ch until we can resume.
	for wg.c >= wg.max {
		wg.c += <-wg.ch
	}
	// Increment our counter and internal waitgroup
	// once our limit falls below our concurrency limit.
	wg.wg.Add(1)
	wg.c++
}

// NewThrottler will return a new WgThrottler with the desired
// maximum concurrency limit 'max'.
func NewThrottler(max int) *WgThrottler {
	return &WgThrottler{
		wg:  sync.WaitGroup{},
		mu:  sync.Mutex{},
		ch:  make(chan int),
		max: max,
		c:   0,
	}
}
