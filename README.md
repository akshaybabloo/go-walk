# go-walk

A directory tree parser

## Instillation

```
go get github.com/akshaybabloo/go-walk
```

## Usage

```go
package main

import (
    "fmt"
    walk "github.com/akshaybabloo/go-walk"
)

func main() {
    dirStats, err := walk.ListDirStat("/", "node_modules")
    if err != nil {
        panic(err)
    }
    fmt.Println(dirStats)
}
```
