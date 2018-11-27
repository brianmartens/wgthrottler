# wgthrottler
If you have a long series of processes that you want to be run on separate goroutines but want to limit the amount of goroutines running at any given time, then this may be what you are looking for. It is as simple as using a sync.WaitGroup but allows you to limit how many goroutines can actually run concurrently.

## Usage

```Go
package main

import(
    "fmt"
    
    "github.com/brianmartens/wgthrottler"
)

func MyFunc(j int) {
    fmt.Println(j)
}

func main() {
    // create a new throttler with a maximum capacity of 5
    th := wgthrottler.NewThrottler(5)

    for i:=0; i<25; i++ {
        // instead of using WaitGroup.Add(1), simply call th.Next()
        th.Next()
        go func(j int) {
            defer th.Done()
            MyFunc(j)
        }(i)
    }

    // Just like sync package's WaitGroup.Wait()
    th.Wait()
}
```
