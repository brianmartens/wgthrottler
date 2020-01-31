package wgthrottler

import (
	"context"
	"sync"
)

// Throttler is an interface which expects three methods: Done(), Wait(), and Next().
// Done() and Wait() should function equivalently to a sync.WaitGroup, whereas Next() blocks until a new goroutine
// may be allocated according to an arbitrary ruleset defined by the implementation.
// Use() starts a session, returning the session as a context.Context.
// This context should be used as the input to Done() and Next() to prevent the case of a deadlock
// whereby one 'user' of the Throttler manages to hoard all capacity in a blocking procedure
type Throttler interface {
	Done(ctx context.Context)
	Wait()
	Next(ctx context.Context)
	Use() context.Context
}

// WgThrottler - A throttled waitgroup for limiting concurrent/parallel processes.
//  cMap - Active count of processes owned by each user of the throttler
//  last - Auto-incrementing integer to use as identifiers for users
//  total - Total utilized concurrency
//  max - Maximum allowed number of active processes
//  ch - Channel used to communicate when a process is complete
type WgThrottler struct {
	sync.Mutex
	cMap  map[int]int
	last  int
	total int
	max   int
	ch    chan struct{}
}

// NewThrottler will return a new WgThrottler with the desired
// maximum concurrency limit 'max'.
func NewThrottler(max int) *WgThrottler {
	return &WgThrottler{
		ch:    make(chan struct{}),
		max:   max,
		total: 0,
		last:  0,
		cMap:  make(map[int]int),
	}
}

// Done is functionally equivalent to a sync.WaitGroup's Done() method.
// An empty struct will be sent through ch and the underlying sync.WaitGroup
func (wg *WgThrottler) Done(ctx context.Context) {
	// get user from context
	u, ok := ctx.Value("user").(int)
	if !ok {
		panic("wg.Next() called with invalid user context. Context must be acquired via a respective call to wg.Use()")
	}

	// release concurrency from the user back to the pool
	wg.dec(u)
}

// Wait is functionally equivalent to a regular sync.WaitGroup's Wait() method.
// This will force the WgThrottler to wait until all running goroutines have completed.
func (wg *WgThrottler) Wait() {
	defer close(wg.ch)
	// wait until total reaches 0
	for range wg.ch {
		if wg.total <= 0 {
			break
		}
	}
}

// Use returns a context to be used in subsequent calls to Next() and Done().
// Use will return nil if the total users already using the throttler meets or exceeds its max concurrency.
func (wg *WgThrottler) Use() context.Context {
	wg.Lock()
	defer wg.Unlock()
	// too many concurrent users given the max level of concurrency
	if len(wg.cMap) >= wg.max {
		return nil
	}
	wg.last++
	wg.cMap[wg.last] = 0
	return context.WithValue(context.Background(), "user", wg.last)
}

// Next will attempt to allocate concurrency from the pool. This will block if the pool is already fully allocated
// or if the user context cannot safely hold more concurrency without risking deadlock
//	ctx := wg.Use()
//  for i := 0; i < 10; i++ {
//    wg.Next(ctx)
//    go func(){
//        defer wg.Done(ctx)
// 		  MyFunc()
//    }
//  }
func (wg *WgThrottler) Next(ctx context.Context) {
	user, ok := ctx.Value("user").(int)
	if !ok {
		panic("wg.Next() called with invalid user context. Context must be acquired via a respective call to wg.Use()")
	}

	// contextMax is used to represent the maximum level of concurrency the user can maintain without the risk of deadlock
	contextMax := wg.max / len(wg.cMap)
	if wg.max%len(wg.cMap) > 0 {
		contextMax++
	}


	if wg.get(user) >= contextMax {
		for range wg.ch {
			if wg.get(user) < contextMax {
				break
			}
		}
	}

	for wg.total >= wg.max {
		<-wg.ch
	}
	wg.inc(user)
}

func (wg *WgThrottler) get(user int) int {
	wg.Lock()
	defer wg.Unlock()
	return wg.cMap[user]
}

func (wg *WgThrottler) inc(user int) int {
	wg.Lock()
	defer wg.Unlock()
	wg.cMap[user]++
	wg.total++
	return wg.cMap[user]
}

func (wg *WgThrottler) dec(user int) int {
	wg.Lock()
	defer wg.Unlock()
	wg.cMap[user]--
	wg.total--
	wg.ch <- struct{}{}
	return wg.cMap[user]
}
