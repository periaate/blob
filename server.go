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
	res.mux.HandleFunc("PUT /b/", preb(res.Set))
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
	w.WriteHeader(http.StatusOK)
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
		HandleErr(w, err)
		return
	}

	buf, err := ensureBody(r)
	if err != nil {
		HandleErr(w, err)
		return
	}

	fp := r.URL.Path
	if err := s.Storage.Add(cType, fp, buf); err != nil {
		if ErrIsBadPath(err) || ErrIsExists(err) || ErrIsIsDir(err) || ErrIsNoDir(err) {
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

func (s *Server) Set(w http.ResponseWriter, r *http.Request) {
	cType, err := getCType(r)
	if err != nil {
		HandleErr(w, err)
		return
	}

	buf, err := ensureBody(r)
	if err != nil {
		HandleErr(w, err)
		return
	}

	fp := r.URL.Path
	n, err := s.Storage.Set(cType, fp, buf)
	if err != nil {
		HandleErr(w, err)
		return
	}

	if n == 0 {
		err = ErrFatal{"no bytes written without error"}
		HandleErr(w, err)
		clog.Fatal("universal error", "err", err.Error(), "fp", fp)
		return
	}

	w.WriteHeader(http.StatusOK)
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

// func (s *Server) Add(w http.ResponseWriter, r *http.Request) {
// 	contentType := r.Header.Get("Content-Type")
// 	enumValue, ok := contentTypeMap[contentType]
// 	if !ok {
// 		http.Error(w, "Unsupported Content-Type", http.StatusUnsupportedMediaType)
// 		return
// 	}
//
// 	fileName, dir, err := s.getFn(r.URL.Path)
// 	switch {
// 	case len(fileName) == 0:
// 		http.Error(w, "no file name", http.StatusBadRequest)
// 		return
// 	case dir:
// 		http.Error(w, "can not create a blob which is a directory", http.StatusBadRequest)
// 		return
// 	case !ErrIsNotFound(err):
// 		http.Error(w, "can not add blob which already exists", http.StatusConflict)
// 		return
// 	case ErrIsNotFound(err):
// 	case err != nil:
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}
//
// 	if err := fsio.EnsureDir(filepath.Dir(fileName)); err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		clog.Fatal("couldn't create blob dir", "err", err)
// 	}
//
// 	fp := filepath.Base(fileName)
// 	clog.Debug("adding blob", "fp", fp, "fn", fileName)
// 	fp = fmt.Sprintf("%d_%s", enumValue, fp)
// 	fp = filepath.Join(filepath.Dir(fileName), fp)
// 	clog.Debug("adding new blob", "fp", fp)
// 	if err := fsio.WriteAll(fp, buf); err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		clog.Fatal("couldn't write blob", "err", err)
// 	}
//
// 	w.WriteHeader(http.StatusCreated)
// }
//
// func (s *Server) Set(w http.ResponseWriter, r *http.Request) {
// 	contentType := r.Header.Get("Content-Type")
// 	enumValue, ok := contentTypeMap[contentType]
// 	if !ok {
// 		http.Error(w, "Unsupported Content-Type", http.StatusUnsupportedMediaType)
// 		return
// 	}
//
// 	fileName, dir, err := s.getFn(r.URL.Path)
// 	switch {
// 	case dir:
// 		http.Error(w, "path to blob is a directory", http.StatusBadRequest)
// 		return
// 	case ErrIsNotFound(err):
// 		http.Error(w, err.Error(), http.StatusNotFound)
// 		return
// 	case err != nil:
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}
//
// 	if err := os.Remove(fileName); err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		clog.Fatal("couldn't remove blob", "err", err)
// 	}
//
// 	fp := filepath.Base(fileName)
// 	blobname, err := SplitBlob(fp)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}
//
// 	fp = blobname[1]
// 	fp = fmt.Sprintf("%d_%s", enumValue, fp)
// 	fp = filepath.Join(filepath.Dir(fileName), fp)
// 	clog.Debug("setting blob", "fp", fp)
// 	if err := fsio.WriteAll(fp, r.Body); err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		clog.Fatal("couldn't write blob", "err", err)
// 	}
//
// 	w.WriteHeader(http.StatusOK)
// }
//
// func (s *Server) Del(w http.ResponseWriter, r *http.Request) {
// 	fileName, dir, err := s.getFn(r.URL.Path)
// 	switch {
// 	case dir:
// 		http.Error(w, "can not delete blob", http.StatusBadRequest)
// 		return
// 	case ErrIsNotFound(err):
// 		http.Error(w, "can not delete file which does not exist", http.StatusNotFound)
// 		return
// 	case err != nil:
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}
//
// 	if err := os.Remove(fileName); err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		clog.Fatal("couldn't remove blob", "err", err)
// 	}
//
// 	w.WriteHeader(http.StatusOK)
// }
//
// func (s *Server) Get(w http.ResponseWriter, r *http.Request) {
// 	fileName, dir, err := s.getFn(r.URL.Path)
// 	switch {
// 	case ErrIsNotFound(err):
// 		clog.Error("couldn't get blob", "err", err)
// 		http.Error(w, err.Error(), http.StatusNotFound)
// 	case err != nil:
// 		http.Error(w, err.Error(), http.StatusBadRequest)
// 	case dir:
// 		files, err := fsio.ReadDir(fileName)
// 		if err != nil {
// 			http.Error(w, err.Error(), http.StatusInternalServerError)
// 			return
// 		}
//
// 		list := make([][2]string, 0, len(files))
// 		for _, f := range files {
// 			fn, err := SplitBlob(f)
// 			if err != nil {
// 				http.Error(w, err.Error(), http.StatusInternalServerError)
// 				return
// 			}
//
// 			list = append(list, fn)
// 		}
//
// 		w.Header().Set("Content-Type", "application/json")
// 		json.NewEncoder(w).Encode(list)
// 	default:
// 		s.ServeFile(w, fileName)
// 	}
// }
//
// func (s *Server) ServeFile(w http.ResponseWriter, fileName string) {
// 	bar, err := os.ReadFile(fileName)
// 	if err != nil {
// 		err = fmt.Errorf("couldn't read blob: %s", err)
// 		http.Error(w, err.Error(), http.StatusNotFound)
// 		return
// 	}
//
// 	res, err := SplitBlob(filepath.Base(fileName))
// 	if err != nil {
// 		err = fmt.Errorf("couldn't split blob: %s", err)
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}
//
// 	w.Header().Set("Content-Type", res[0])
// 	w.Write(bar)
// }
//
// func (s *Server) getFn(fp string) (res string, isDir bool, err error) {
// 	if len(fp) == 0 {
// 		err = ErrBadPath{
// 			fp,
// 			fmt.Errorf("empty path"),
// 		}
// 		clog.Debug("empty path", "fp", fp)
// 		return
// 	}
//
// 	clog.Debug("getFn", "fp", fp)
// 	res = "./" + filepath.Join(s.Root, fp[1:])
// 	if isDir = fsio.IsDir(res); isDir {
// 		clog.Debug("filepath is dir, returning", "res", res)
// 		return
// 	}
//
// 	clog.Debug("filepath is not dir, checking if exists", "res", res)
//
// 	if fsio.Exists(res) {
// 		clog.Warn("requests shouldn't be made directly", "path", res)
// 		return
// 	}
//
// 	base := filepath.Base(res)
//
// 	dir := filepath.Dir(res)
// 	if !fsio.IsDir(dir) {
// 		err = ErrBadPath{
// 			fp,
// 			fmt.Errorf("path points to a file within a file"),
// 		}
// 		clog.Debug("path points to a file within a file", "dir", dir, "base", base)
// 		return
// 	}
//
// 	res = gen.First(str.HasSuffix(base))(gen.Must(fsio.ReadDir(dir)))
// 	if len(res) == 0 {
// 		err = ErrNotExists{
// 			fp,
// 			fmt.Errorf("no file found"),
// 		}
// 		clog.Debug("no file found", "dir", dir)
// 		return
// 	}
// 	res = fsio.Normalize(res)
//
// 	clog.Debug("found file", "res", res)
// 	return
// }
