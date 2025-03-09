package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/rocketssan/toolkit"
)

func main() {
	mux := routes()

	log.Println("starting server on port 8081")

	err := http.ListenAndServe(":8081", mux)
	if err != nil {
		log.Fatal(err)
	}
}

func routes() http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("."))))
	mux.HandleFunc("/upload", uploadFiles)
	mux.HandleFunc("/upload-one", uploadFile)

	return mux
}

func uploadFiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	t := toolkit.Tools{
		MaxFileSize: 1024 * 1024 * 1024,
	}

	uploadFiles, err := t.UploadFiles(r, "./uploads", []string{"image/jpeg", "image/png"})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	out := ""
	for _, file := range uploadFiles {
		out += fmt.Sprintf("Uploaded %s to the uploads folder, renamed to %s\n", file.NewFileName, file.OriginalFileName)
	}

	_, _ = w.Write([]byte(out))
}

func uploadFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	t := toolkit.Tools{
		MaxFileSize: 1024 * 1024 * 1024,
	}

	f, err := t.UploadOneFile(r, "./uploads", []string{"image/jpeg", "image/png"})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, _ = w.Write([]byte(fmt.Sprintf("Uploaded 1 file, %s to the uploads folder", f.OriginalFileName)))
}
