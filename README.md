# Go Daemon

A daemon package for use with Golang.

## Usage

1. Import package `github.com/miaia/daemon`
2. Add `daemon.RunDaemon()` to your init function in the main.go of the project
3. Use command `./{Binary file} start|stop|install|uninstall`

## Example

The following is a simple http server.

``` Go
package main

import (
    "fmt"
    "net/http"

    "github.com/miaia/daemon"
)

func init() {
    daemon.RunDaemon()
}

func handle(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hello World!")
}

func main() {
    http.HandleFunc("/", handle)
    err := http.ListenAndServe(":9090", nil)
    if err != nil {
        fmt.Println(err)
    }
}
```
