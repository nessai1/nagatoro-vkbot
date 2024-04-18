package main

import (
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
		writer.Write([]byte("good request!"))
		log.Println("got new request!")
	})
	http.ListenAndServe("0.0.0.0:3232", nil)
}
