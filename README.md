# Time Wheel
support massive scheduled tasks in short time, high performance

## Install
```go
go get -u github.com/pzs-pzs/timingwheel
```

## Example

```go
package main

import (
	"github.com/pzs-pzs/timingwheel"
	"time"
)

func main() {
	tw, err := timingwheel.NewTimingWheel(1*time.Second, 10)
	if err != nil {
		panic(err)
	}
	beginTime := time.Now().Unix()
	err = tw.AddOnceTask(5*time.Second, "taskId", false, func(key string) {
		println(key)
		endTime := time.Now().Unix()
		println(endTime - beginTime)
	})
	if err != nil {
		panic(err)
	}
	tw.Start()
	time.Sleep(20 * time.Second)
	tw.Stop()

}
```
