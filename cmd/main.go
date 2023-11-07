package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

func main() {
	fmt.Println("Begin")

	mux := http.NewServeMux()
	mux.HandleFunc("/", getRoot)
	mux.HandleFunc("/hi", getHi)

	err := http.ListenAndServe(":3333", mux)
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}

func getRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Println("got request")
	_, err := io.WriteString(w, "This is root of my webserver\n")
	if err != nil {
		log.Fatal(err)
	}
}
func getHi(w http.ResponseWriter, r *http.Request) {
	fmt.Println("hi request")
	_, err := io.WriteString(w, "This is hi route of my webserver\n")
	if err != nil {
		log.Fatal(err)
	}
}
