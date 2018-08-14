package main

import (
	"fmt"
	"log"
	"net/http"
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Responding to %s!", r.URL.Path[1:])
}

func main() {
	http.HandleFunc("/", handler)
	log.Println("Started server :8080..")
	log.Fatal(http.ListenAndServe("localhost:8080", nil))
}
