package main

import (
    "fmt"
    "net/http"
)

func main() {
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintln(w, "Hello World")
    })

    fmt.Println("Reporting service listening on :8007")

    http.ListenAndServe(":8007", nil)
}