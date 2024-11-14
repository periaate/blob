package blob

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

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
		Storage: &Storage{Root: root},
		mux:     http.NewServeMux(),
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
	if err := s.Storage.Mkdir(r.URL.Path); err != nil {
		HandleErr(w, err)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (s *Server) DirGet(w http.ResponseWriter, r *http.Request) {
	blobs, err := s.Storage.Lsdir(r.URL.Path)
	if err != nil {
		HandleErr(w, err)
		return
	}

	bar, err := json.Marshal(blobs)
	if err != nil {
		HandleErr(w, err)
		return
	}
	if len(blobs) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(bar)
}

func (s *Server) DirDel(w http.ResponseWriter, r *http.Request) {
	if err := s.Storage.Rmdir(r.URL.Path); err != nil {
		HandleErr(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) { s.mux.ServeHTTP(w, r) }

func getCType(r *http.Request) (res CType, err error) {
	contentType := r.Header.Get("Content-Type")
	if contentType == "" {
		err = ErrBadRequest{"no Content-Type header"}
		return
	}

	res = ContentType(contentType)
	return
}

func ensureBody(r *http.Request) (buf *bytes.Buffer, err error) {
	// is there a better way to do this?
	buf = bytes.NewBuffer([]byte{})
	n, err := io.Copy(buf, r.Body)
	if err != nil {
		err = ErrBadRequest{err.Error()}
		return
	}
	defer r.Body.Close()

	if n == 0 {
		return nil, ErrBadRequest{"empty body"}
	}

	return
}

func (s *Server) Add(w http.ResponseWriter, r *http.Request) {
	cType, err := getCType(r)
	if err != nil {
		clog.Error("failed to get content type", "err", err)
		HandleErr(w, err)
		return
	}

	buf, err := ensureBody(r)
	if err != nil {
		clog.Error("failed to ensure body", "err", err)
		HandleErr(w, err)
		return
	}

	fp := r.URL.Path
	if err := s.Storage.Add(cType, fp, buf); err != nil {
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
	http.Error(w, err.Error(), StatusOfErr(err))
	clog.Error(err.Error())
}

func (s *Server) Del(w http.ResponseWriter, r *http.Request) {
	fp := r.URL.Path
	if err := s.Storage.Del(fp); err != nil {
		HandleErr(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Server) Get(w http.ResponseWriter, r *http.Request) {
	fp := r.URL.Path
	rc, cType, err := s.Storage.Get(fp)
	if err != nil {
		HandleErr(w, err)
		return
	}
	defer rc.Close()

	w.Header().Set("Content-Type", cType.String())
	io.Copy(w, rc)
}
