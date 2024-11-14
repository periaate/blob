package wrap

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/periaate/blob"
	"github.com/periaate/blume/clog"
	"github.com/periaate/blume/fsio"
	"github.com/periaate/blume/str"
)

var _ blob.Storager = &Storage{}

func New(base string, vRoot string) (s *Storage, err error) {
	u, err := url.Parse(base)
	if err != nil {
		return
	}

	clog.Debug("new storage", "base", u.String(), "vRoot", vRoot)

	s = &Storage{
		vRoot: vRoot,
		base:  u,
	}
	return
}

type Storage struct {
	// vRoot defines a virtual root for the storage.
	// It is used to prefix all paths in the storage.
	// Default is empty string, which means no prefix.
	vRoot string
	base  *url.URL
}

func (s *Storage) req(bPath string, isDir bool, method string, r io.Reader) (rq *http.Request, err error) {
	if str.Contains("..")(bPath) {
		err = blob.ErrIllegalPath{Path: bPath}
		return
	}

	sub := "b"
	if isDir {
		sub = "d"
	}
	fmt.Println("base", s.base.String())
	fmt.Println("sub", sub)
	fmt.Println("vRoot", s.vRoot)
	fmt.Println("bPath", bPath)

	uri := fsio.Join(s.base.String(), sub, s.vRoot, bPath)
	fmt.Println("uri", uri)
	fmt.Println("base", s.base.String())
	rq, err = http.NewRequest(method, uri, r)
	return
}

func Read(r io.Reader) (msg string) {
	buf := make([]byte, 1024)
	n, _ := r.Read(buf)
	msg = string(buf[:n])
	return
}

func (s *Storage) Add(bType blob.CType, bPath string, r io.Reader) (err error) {
	req, err := s.req(bPath, false, http.MethodPost, r)
	if err != nil {
		clog.Error("failed to make request", "err", err)
		return
	}

	req.Header.Set("Content-Type", bType.String())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		clog.Error("failed to send request", "err", err)
		return
	}

	msg := Read(resp.Body)

	err = StatusToErr(resp.StatusCode, "")
	if err != nil {
		clog.Error("failed to add", "err", err, "status", resp.StatusCode, "path", req.URL.String(), "method", req.Method, "type", bType, "msg", msg)
		return
	}

	return
}

func (s *Storage) Del(bPath string) (err error) {
	req, err := s.req(bPath, false, http.MethodDelete, nil)
	if err != nil {
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		clog.Error("failed to delete", "err", err)
		return
	}

	err = StatusToErr(resp.StatusCode, "")
	if err != nil {
		return
	}

	return
}

func (s *Storage) Get(bPath string) (r io.ReadCloser, err error) {
	req, err := s.req(bPath, false, http.MethodGet, nil)
	if err != nil {
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		clog.Error("failed to get", "err", err)
		return
	}

	err = StatusToErr(resp.StatusCode, "")
	if err != nil {
		return
	}

	r = resp.Body
	return
}

func (s *Storage) Mkdir(dPath string) (err error) {
	req, err := s.req(dPath, true, http.MethodPost, nil)
	if err != nil {
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}

	err = StatusToErr(resp.StatusCode, "")
	return
}

func StatusToErr(status int, msg string) error {
	switch status {
	case http.StatusCreated:
		return nil
	case http.StatusOK:
		return nil
	case http.StatusNoContent:
		return nil
	case http.StatusNotFound:
		return blob.ErrBadPath{}
	case http.StatusConflict:
		return blob.ErrExists{}
	case http.StatusBadRequest:
		return blob.ErrBadRequest{}
	case http.StatusForbidden:
		return blob.ErrBadPath{}
	default:
		return blob.ErrBadRequest{}
	}
}

func (s *Storage) RmDir(dPath string) (err error) {
	req, err := s.req(dPath, true, http.MethodDelete, nil)
	if err != nil {
		return
	}

	_, err = http.DefaultClient.Do(req)
	return
}

func (s *Storage) Lsdir(dPath string) (blobs [][2]string, err error) {
	req, err := s.req(dPath, true, http.MethodGet, nil)
	if err != nil {
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	err = StatusToErr(resp.StatusCode, "")
	if err != nil {
		return
	}

	err = json.NewDecoder(resp.Body).Decode(&blobs)
	return
}
