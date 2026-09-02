package main

import (
    "fmt"
    "net/http"
)

func main() {
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintln(w, "Hello World")
    })

    fmt.Println("Bookkeeping service listening on :8005")

    http.ListenAndServe(":8005", nil)
}