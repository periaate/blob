package blob

import (
	"encoding/json"
	"io"
	"net/http"
	"sync"

	"github.com/periaate/blume/clog"
	"github.com/periaate/blume/x/hnet"
)

// Server struct implements http.Handler and fs.FS
type Server struct {
	*Storage
	mux *http.ServeMux
}

// New creates a new Storage instance
func New(root string) *Server {
	res := &Server{
		Storage: &Storage{
			Root:  root,
			Cache: map[string]*Blob{},
			mut:   sync.Mutex{},
		},
		mux: http.NewServeMux(),
	}
	preb := hnet.Pre("/b/")
	pred := hnet.Pre("/d/")

	res.mux.HandleFunc("POST /b/", preb(res.Add))
	res.mux.HandleFunc("GET /b/", preb(res.Get))
	res.mux.HandleFunc("DELETE /b/", preb(res.Del))

	res.mux.HandleFunc("POST /d/", pred(res.DirAdd))
	res.mux.HandleFunc("GET /d/", pred(res.DirGet))
	res.mux.HandleFunc("DELETE /d/", pred(res.DirDel))
	return res
}

func (s *Server) DirAdd(w http.ResponseWriter, r *http.Request) {
	if err := s.Storage.MkDir(r.URL.Path); err != nil {
		HandleErr(w, err)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (s *Server) DirGet(w http.ResponseWriter, r *http.Request) {
	s.Storage.WithBlob(func(b *Blob, err error) {
		if err != nil {
			HandleErr(w, err)
			return
		}

		blobs, err := b.Ls()
		if err != nil {
			HandleErr(w, err)
			return
		}

		bar, err := json.Marshal(blobs)
		if err != nil {
			HandleErr(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(bar)
	})(r.URL.Path)
}

func (s *Server) DirDel(w http.ResponseWriter, r *http.Request) {
	s.Storage.WithBlob(func(b *Blob, err error) {
		if err == nil {
			err = s.Storage.RmDir(b)
		}
		HandleErr(w, err)
	})
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) { s.mux.ServeHTTP(w, r) }

func (s *Server) Add(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType == "" {
		HandleErr(w, ErrBadRequest{"no Content-Type header"})
		return
	}

	ct := GetCT(contentType)

	fp := r.URL.Path
	if err := s.Storage.Add(fp, r.Body, ct); err != nil {
		if ErrIs[ErrBadPath](err) ||
			ErrIs[ErrExists](err) ||
			ErrIs[ErrIsDir](err) ||
			ErrIs[ErrNoDir](err) {
			err = ErrBadRequest{err.Error()}
		}

		HandleErr(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func HandleErr(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}
	http.Error(w, err.Error(), StatusOfErr(err))
	clog.Error(err.Error())
}

func (s *Server) Del(w http.ResponseWriter, r *http.Request) {
	s.Storage.WithBlob(func(b *Blob, err error) {
		if err == nil {
			err = s.Storage.Del(b)
		}
		HandleErr(w, err)
	})(r.URL.Path)
}

func (s *Server) Get(w http.ResponseWriter, r *http.Request) {
	s.Storage.WithBlob(func(b *Blob, err error) {
		if err == nil {
			w.Header().Set("Content-Type", b.Type.String())
			_, err = io.Copy(w, b)
		}
		HandleErr(w, err)
	})(r.URL.Path)
}
