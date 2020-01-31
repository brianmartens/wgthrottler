# wgthrottler
A simple, easy-to-use waitgroup that limits concurrency across any number of running goroutines without the risk of deadlock


## Usage

```Go
package main

import(
    "fmt"
    "time"
    
    "github.com/brianmartens/wgthrottler"
)


func main() {
    // To begin, declare a new throttler with a fixed level of concurrency.
	th := wgthrottler.NewThrottler(5)
    // Create some user sessions. We can create up to 5 in this case, but let's go with 3.
	user1, user2, user3 := th.Use(), th.Use(), th.Use()
	// Run countdowns for each user concurrently
	go userCountdown(user1, th)
	go userCountdown(user2, th)
	go userCountdown(user3, th)
    // Wait until done...
	th.Wait()
	fmt.Println("Done!")
}

// Uses the throttler along with the given user context to safely countdown concurrently with other active users
func userCountdown(user context.Context, th *wgthrottler.WgThrottler) {
	for i := 0; i < 10; i++ {
		th.Next(user)
		go func(j int) {
			defer th.Done(user)
			time.Sleep(200 * time.Millisecond)
			fmt.Println("user:", user.Value("user"), j)
		}(i)
	}
}

```
