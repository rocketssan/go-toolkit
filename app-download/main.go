package main

import (
	"log"
	"net/http"

	"github.com/rocketssan/toolkit"
)

func main() {
	mux := routes()

	err := http.ListenAndServe(":8081", mux)
	if err != nil {
		log.Fatal(err)
	}
}

func routes() http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("."))))
	mux.HandleFunc("/download", downloadFile)

	return mux
}

func downloadFile(w http.ResponseWriter, r *http.Request) {
	t := toolkit.Tools{}

	t.DownloadStaticFile(w, r, "./files", "hacker.png", "person.png")
}
