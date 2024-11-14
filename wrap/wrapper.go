package wrap

import (
	"encoding/json"
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

func (s *Storage) req(bPath string, isDir bool, method string) (r *http.Request, err error) {
	if str.Contains("..")(bPath) {
		err = blob.ErrIllegalPath{Path: bPath}
		return
	}

	var uri string

	if isDir {
		uri = fsio.Join("d", s.vRoot, bPath, "/")
		uri = s.base.String() + "/" + uri
	} else {
		uri = fsio.Join("b", s.vRoot, bPath)
		uri = s.base.String() + "/" + uri
	}

	reqUrl, err := url.ParseRequestURI(uri)
	if err != nil {
		err = blob.ErrBadPath{Path: bPath}
		return
	}

	r, err = http.NewRequest(method, reqUrl.String(), nil)
	return
}

func (s *Storage) Add(bType blob.CType, bPath string, r io.Reader) (err error) {
	req, err := s.req(bPath, false, http.MethodPost)
	if err != nil {
		return
	}

	req.Body = io.NopCloser(r)
	req.Header.Set("Content-Type", bType.String())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}

	err = StatusToErr(resp.StatusCode, "")
	if err != nil {
		return
	}

	return
}

func (s *Storage) Del(bPath string) (err error) {
	req, err := s.req(bPath, false, http.MethodDelete)
	if err != nil {
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}

	err = StatusToErr(resp.StatusCode, "")
	if err != nil {
		return
	}

	return
}

func (s *Storage) Get(bPath string) (r io.ReadCloser, err error) {
	req, err := s.req(bPath, false, http.MethodGet)
	if err != nil {
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
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
	req, err := s.req(dPath, true, http.MethodPost)
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
	req, err := s.req(dPath, true, http.MethodDelete)
	if err != nil {
		return
	}

	_, err = http.DefaultClient.Do(req)
	return
}

func (s *Storage) Lsdir(dPath string) (blobs [][2]string, err error) {
	req, err := s.req(dPath, true, http.MethodGet)
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
