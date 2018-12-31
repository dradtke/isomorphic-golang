package main

import (
	"log"
	"net/http"
)

func main() {
	_, err := http.Dir("./static").Open("index.js")
	if err != nil {
		log.Fatal(err)
	}
}
