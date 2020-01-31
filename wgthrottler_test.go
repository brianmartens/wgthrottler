package wgthrottler

import (
	"context"
	"testing"
	"time"
)

func TestThrottle(t *testing.T) {
	th := NewThrottler(5)
	user1, user2, user3 := th.Use(), th.Use(), th.Use()
	go userCountdown(user1, th, t)
	go userCountdown(user2, th, t)
	go userCountdown(user3, th, t)

	th.Wait()
	t.Log("Done!")
}

func userCountdown(user context.Context, th *WgThrottler, t *testing.T) {
	for i := 0; i < 10; i++ {
		th.Next(user)
		go func(j int) {
			defer th.Done(user)
			time.Sleep(200 * time.Millisecond)
			t.Log("user:", user.Value("user"), j)
		}(i)
	}
}
