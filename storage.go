package blob

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/periaate/blume/clog"
	"github.com/periaate/blume/fsio"
	"github.com/periaate/blume/str"
)

type Blob struct {
	Name  string
	IsDir bool
	Type  ContentType
	mut   sync.Mutex
}

// io.Reader
func (b *Blob) Read(p []byte) (n int, err error) {
	b.mut.Lock()
	defer b.mut.Unlock()
	f, err := os.Open(b.Name)
	if err != nil {
		return
	}
	defer f.Close()
	return f.Read(p)
}

// io.Writer
func (b *Blob) Write(p []byte) (n int, err error) {
	b.mut.Lock()
	defer b.mut.Unlock()
	f, err := os.Create(b.Name)
	if err != nil {
		return
	}
	defer f.Close()
	return f.Write(p)
}

func (b *Blob) Delete() (err error) {
	if b.IsDir {
		return ErrIsDir{b.Name}
	}
	b.mut.Lock()
	defer b.mut.Unlock()
	if fsio.IsDir(b.Name) {
		return ErrIsDir{b.Name}
	}
	return os.Remove(b.Name)
}

func (b *Blob) Rm() (err error) {
	if !b.IsDir {
		return ErrDirNotEmpty{b.Name}
	}
	b.mut.Lock()
	defer b.mut.Unlock()
	if !fsio.IsDir(b.Name) {
		return ErrDirNotEmpty{b.Name}
	}
	files, err := fsio.ReadDir(b.Name)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return ErrDirNotEmpty{Path: b.Name}
	}
	return os.Remove(b.Name)
}

func (b *Blob) Ls() (blobs [][2]string, err error) {
	if !b.IsDir {
		err = ErrDirNotEmpty{b.Name}
		return
	}
	b.mut.Lock()
	defer b.mut.Unlock()
	if !fsio.IsDir(b.Name) {
		err = ErrDirNotEmpty{b.Name}
		return
	}
	files, err := fsio.ReadDir(b.Name)
	if err != nil {
		return
	}

	var blob [2]string
	for _, p := range files {
		blob, err = SplitBlob(fsio.Base(p))
		if err != nil {
			return
		}

		blobs = append(blobs, blob)
	}

	return
}

func (s *Storage) Del(b *Blob) (err error) {
	s.mut.Lock()
	defer s.mut.Unlock()
	err = b.Delete()
	if err != nil {
		return
	}
	delete(s.Cache, b.Name)
	return
}

func (s *Storage) RmDir(b *Blob) (err error) {
	s.mut.Lock()
	defer s.mut.Unlock()
	err = b.Rm()
	if err != nil {
		return
	}
	delete(s.Cache, b.Name)
	return
}

type Storage struct {
	Root  string
	Cache map[string]*Blob
	mut   sync.Mutex
}

/*
blob
	add: add new blob
	del: delete existing blob
	get: get existing blob
*/

func (s *Storage) MkDir(bPath string) (err error) {
	s.mut.Lock()
	defer s.mut.Unlock()
	_, err = s.FindBlobMutless(bPath)
	if err == nil {
		return ErrExists{Path: bPath}
	}
	err = nil

	fp := fsio.Join(s.Root, bPath)
	err = fsio.EnsureDir(fp)
	if err != nil {
		return
	}

	s.Cache[fp] = &Blob{
		Name:  fp,
		IsDir: true,
		mut:   sync.Mutex{},
	}
	return
}

func (s *Storage) Add(bPath string, r io.Reader, ct ContentType) (err error) {
	s.mut.Lock()
	defer s.mut.Unlock()
	_, err = s.FindBlobMutless(bPath)
	if err == nil {
		return ErrExists{Path: bPath}
	}
	err = nil

	bp := fsio.Join(s.Root, bPath)
	base := fsio.Base(bp)
	dir := fsio.Dir(bp)
	bp = fsio.Join(dir, fmt.Sprintf("%d_%s", ct, base))

	err = fsio.WriteNew(bp, r)
	if os.IsNotExist(err) {
		return ErrNoDir{Path: bPath}
	}

	if err != nil {
		return
	}

	s.Cache[bp] = &Blob{bp, false, ct, sync.Mutex{}}
	return
}

func (s *Storage) WithBlob(fn func(b *Blob, err error)) func(path string) {
	return func(path string) {
		var b *Blob
		if str.Contains("..")(path) {
			fn(b, ErrIllegalPath{path})
			return
		}

		bp := fsio.Join(s.Root, path)
		fn(s.FindBlob(bp))
	}
}

func (s *Storage) FindBlobMutless(bp string) (b *Blob, err error) {
	b, ok := s.Cache[bp]
	switch {
	case !ok:
		return nil, ErrNotExists{}
	default:
		return b, nil
	}
}

func (s *Storage) FindBlob(bp string) (b *Blob, err error) {
	s.mut.Lock()
	defer s.mut.Unlock()
	return s.FindBlobMutless(bp)
}

/*
tree
	mkdir: make new directory
	rmdir: delete existing directory
	lsdir: list blobs in directory
*/

func SplitBlob(path string) (res [2]string, err error) {
	parts := strings.Split(path, "_")
	if len(parts) < 2 {
		err = fmt.Errorf("blob path does not have valid content type %s", path)
		return
	}
	if len(parts) > 2 {
		parts[1] = strings.Join(parts[1:], "_")
	}

	enumValue, err := strconv.Atoi(parts[0])
	if err != nil {
		err = fmt.Errorf("invalid blob path %s", path)
		return
	}

	contentType := ContentType(enumValue).String()

	res[0] = contentType
	res[1] = parts[1]
	clog.Debug("split blob", "type", contentType, "name", parts[1])
	return
}
