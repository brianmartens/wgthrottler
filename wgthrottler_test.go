package wgthrottler_test

import (
	"fmt"
	"testing"

	"github.com/brianmartens/wgthrottler"
)

func TestMain(M *testing.M) {
	fmt.Println("Launching wgthrottler test")
	M.Run()
}

func TestThrottle(t *testing.T) {
	th := wgthrottler.NewThrottler(3)
	for i := 0; i < 10; i++ {
		th.Next()
		go func(j int) {
			defer th.Done()
			t.Log(j)
		}(i)
	}
	th.Wait()
	t.Log("Test successful")
}
