package main

import (
    "fmt"
    "net/http"
)

func main() {
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintln(w, "Hello World")
    })

    fmt.Println("Auth service listening on :8003")

    http.ListenAndServe(":8003", nil)
}