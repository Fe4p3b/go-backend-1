package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"time"
)

type UploadHandler struct {
	HostAddr  string
	UploadDir string
}

func main() {
	uploadHandler := &UploadHandler{
		UploadDir: "upload",
		HostAddr:  "localhost:8080",
	}

	http.Handle("/upload", uploadHandler)
	srv := &http.Server{
		Addr:         ":8000",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	dirToServe := http.Dir(uploadHandler.UploadDir)
	fs := &http.Server{
		Addr:         ":8080",
		Handler:      http.FileServer(dirToServe),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		log.Fatal(srv.ListenAndServe())
	}()

	log.Fatal(fs.ListenAndServe())
}

func (h *UploadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.uploadGetHandler(w, r)
	case http.MethodPost:
		h.uploadPostHandler(w, r)
	}
}

func (h *UploadHandler) uploadGetHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("ext")
	files, err := ioutil.ReadDir(h.UploadDir)
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range files {
		name := f.Name()
		ext := filepath.Ext(name)
		if q == "" || q == ext {
			fmt.Fprintf(w, "%s %d %s\n", name, f.Size(), ext)
		}
	}
}

func (h *UploadHandler) uploadPostHandler(w http.ResponseWriter, r *http.Request) {
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Unable to read file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		http.Error(w, "Unable to read file", http.StatusBadRequest)
		return
	}

	filePath := h.UploadDir + "/" + header.Filename

	err = ioutil.WriteFile(filePath, data, 0777)
	if err != nil {
		log.Println(err)
		http.Error(w, "Unable to save file", http.StatusInternalServerError)
		return
	}

	fileLink := h.HostAddr + "/" + header.Filename
	fmt.Fprintln(w, fileLink)
}
