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
	http.HandleFunc("/abc", handler)
	log.Println("Started server :8123..")
	log.Fatal(http.ListenAndServe(":8123", nil))
}
