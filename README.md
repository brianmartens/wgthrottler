# wgthrottler
A throttled waitgroup implementation for launching parallel processes.

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
