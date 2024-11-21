package blob

import (
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/periaate/blume/clog"
	"github.com/periaate/blume/fsio"
	"github.com/periaate/blume/gen"
	"github.com/periaate/blume/str"
)

type Bucket struct {
	Name string
	V    map[string]string
	mut  sync.Mutex
}

// Filepath path to Blobpath
func FtoB(fp string) (bucket, name string, ct ContentType) {
	fps := strings.TrimFunc(fp, gen.Is('\\', '/', '.'))
	sar := str.SplitWithAll(fps, false, "/")
	if len(sar) < 2 {
		return
	}
	bucket = sar[len(sar)-2]
	name = sar[len(sar)-1]
	s := str.SplitWithAll(name, false, "_")
	if len(s) > 1 {
		ct = GetCT(s[0])
		fmt.Println("ab", s[0], "ct", ct, "s", s[1:])
		name = strings.Join(s[1:], "_")
	}
	clog.Debug("FtoB", "input", fp, "bucket", bucket, "name", name, "ct", ct)

	return
}

// Blobpath to Filepath
func BtoF(bucket, name string, ct ContentType) (fp string) {
	fp = fmt.Sprintf("%s/%v_%s", bucket, ct, name)
	clog.Debug("BtoF", "fp", fp)
	return
}

type ErrBadRequest struct{ msg string }

func (e ErrBadRequest) Error() string { return e.msg }

type Storage struct {
	Root string
	m    map[string]*Bucket
}

func NewStorage(root string) (s *Storage, err error) {
	if root == "" {
		return
	}

	err = fsio.EnsureDir(root)
	if err != nil {
		return
	}

	s = &Storage{
		Root: root,
		m:    make(map[string]*Bucket),
	}

	res, err := fsio.ReadDir(root)
	if err != nil {
		return
	}

	for _, re := range res {
		if !str.HasSuffix("/")(re) {
			clog.Debug("found file in buckets", "fp", re)
			continue
		}
		sar := str.SplitWithAll(re, false, "/")

		bname, ok := gen.GetPop(sar)
		if !ok {
			clog.Error("failed to get bucket name", "re", re)
			continue
		}
		clog.Debug("got bucket name", "bucket", bname)

		buck := &Bucket{
			Name: bname,
			V:    make(map[string]string),
		}

		s.m[bname] = buck
		err = buck.Refresh(root)
		if err != nil {
			clog.Error("refresh failed", "err", err)
			return
		}
	}

	return
}

func (b *Bucket) Refresh(root string) (err error) {
	b.mut.Lock()
	defer b.mut.Unlock()
	fp := fsio.Join(root, b.Name)
	bres, err := fsio.ReadDir(fp)
	if err != nil {
		clog.Error("read dir failed", "err", err)
		return
	}

	b.V = make(map[string]string, len(bres))

	clog.Debug("found bucket", "bucket", fp)
	for _, bre := range bres {
		bucket, blob, ct := FtoB(bre)
		b.V[blob] = BtoF(bucket, blob, ct)
	}
	return
}

func (s *Storage) Get(path string) (fp string, err error) {
	bucket, name, _ := FtoB(path)
	buck, ok := s.m[bucket]
	if !ok {
		err = ErrBadRequest{msg: "bucket doesn't exist"}
		return
	}

	fp, ok = buck.V[name]
	if !ok {
		err = ErrBadRequest{msg: "blob doesn't exist"}
		return
	}

	return
}

func (s *Storage) Set(path string, ct ContentType, r io.Reader) (err error) {
	bucket, name, _ := FtoB(path)
	buck, ok := s.m[bucket]
	if !ok {
		err = ErrBadRequest{msg: "bucket doesn't exist"}
		return
	}

	fp := BtoF(bucket, name, ct)
	fp = fsio.Join(s.Root, fp)
	err = fsio.WriteAll(fp, r)
	if err != nil {
		return
	}

	buck.V[name] = fp
	return
}
