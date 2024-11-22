package server

import (
	"io"
	"net/http"

	"github.com/periaate/blob"
	"github.com/periaate/blume/er"
	"github.com/periaate/blume/fsio"
)

func NewServer() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{bucket}/{name}", Get)
	mux.HandleFunc("POST /{bucket}/{name}", Set)
	mux.HandleFunc("DELETE /{bucket}/{name}", Del)

	return mux
}

func GetContentType(r *http.Request) (ct blob.ContentType, nerr er.Net) {
	cth := r.Header.Get("Content-Type")
	if cth == "" {
		nerr = er.BadRequest{Msg: "Content-Type header missing"}
		return
	}

	ct = blob.GetCT(cth)
	return
}

func PathValues(r *http.Request) (bucket, name string, err er.Net) {
	bucket = r.PathValue("bucket")
	name = r.PathValue("name")
	if bucket == "" || name == "" {
		err = er.BadRequest{Msg: "bucket or blob empty"}
	}
	return
}

func Get(w http.ResponseWriter, r *http.Request) {
	bucket, name, nerr := PathValues(r)
	if nerr != nil {
		http.Error(w, nerr.Error(), nerr.Status())
		return
	}

	reader, ct, err := blob.Blob(fsio.Join(bucket, name)).Get()
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", ct.String())
	_, err = io.Copy(w, reader)
	if err != nil {
		http.Error(w, "couldn't write blob to response", http.StatusInternalServerError)
	}
}

func Set(w http.ResponseWriter, r *http.Request) {
	bucket, name, nerr := PathValues(r)
	if nerr != nil {
		http.Error(w, nerr.Error(), nerr.Status())
		return
	}

	ct, nerr := GetContentType(r)
	if nerr != nil {
		http.Error(w, nerr.Error(), nerr.Status())
		return
	}

	err := blob.Blob(fsio.Join(bucket, name)).Set(r.Body, ct)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func Del(w http.ResponseWriter, r *http.Request) {
	bucket, name, nerr := PathValues(r)
	if nerr != nil {
		http.Error(w, nerr.Error(), nerr.Status())
		return
	}

	err := blob.Blob(fsio.Join(bucket, name)).Del()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
