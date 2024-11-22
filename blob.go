package blob

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/periaate/blume/clog"
	"github.com/periaate/blume/er"
	"github.com/periaate/blume/fsio"
	"github.com/periaate/blume/gen"
	"github.com/periaate/blume/str"
)

/*
Type: {Content-Type}
Blobpath: {bucket}/{blob}
Blobname: {Type}{blob}
Filepath: {root}/{bucket}/{Blobname}


Type is two base64 letters.

Content-Types are provided during a Set operation on the Index.

*/

func SetIndex(root string) {
	fsio.EnsureDir(root)
	i = &Index{
		SyncMap: gen.NewSyncMap[Blob, ContentType](),
		Root:    root,
	}

	// Load all blobs from the root directory.
	filepaths, _ := fsio.ReadDir(root)
	for _, filepath := range filepaths {
		fmt.Println(filepath)
		if !fsio.IsDir(filepath) {
			clog.Warn("bucket is not a directory", "blob", filepath)
			continue
		}

		bucket := fsio.Base(filepath)
		blobpaths, _ := fsio.ReadDir(filepath)
		for _, blobpath := range blobpaths {
			fmt.Println(blobpath)
			if fsio.IsDir(blobpath) {
				clog.Warn("blob is a directory", "blob", blobpath)
				continue
			}

			base := fsio.Base(blobpath)
			ct := GetCT(base[:2])
			if ct == -1 {
				clog.Warn("invalid content type", "blob", blobpath)
				continue
			}

			name := base[2:]

			blob := Blob(fsio.Join(bucket, name))
			i.Set(blob, ct)
			if _, ok := i.Get(blob); !ok {
				clog.Warn("blob not set in index", "blob", blob)
			}
		}
	}
}

func CloseIndex() {
	i = nil
}

var i *Index

type Index struct {
	*gen.SyncMap[Blob, ContentType]
	Root string
	mut  sync.RWMutex
}

type Blob string

func (b Blob) Split() (bucket, blob string, err error) {
	sar := str.SplitWithAll(string(b), false, "/")
	if len(sar) != 2 {
		err = er.InvalidData{
			Msg:     "Blob has invalid format: [" + string(b) + "]",
			Has:     "split didn't have 2 results",
			Expects: "Blob format is: {bucket}/{blob}",
		}
		return
	}

	bucket = sar[0]
	blob = sar[1]
	return
}

func Filepath(b Blob, ct ContentType) (res string, err error) {
	bucket, blob, err := b.Split()
	if err != nil {
		return
	}
	res = fsio.Join(i.Root, bucket, ct.Fmt()+blob)
	return
}

func (b Blob) File() (res string, ct ContentType, err error) {
	ct, err = b.Type()
	if err != nil {
		return
	}

	bucket, blob, err := b.Split()
	if err != nil {
		return
	}

	res = fsio.Join(i.Root, bucket, ct.Fmt()+blob)
	return
}

func (b Blob) Type() (res ContentType, err error) {
	res, ok := i.Get(b)
	if !ok {
		err = er.NotFound{
			Requested: "Blob",
			From:      "Index",
			With:      string(b),
			Msg:       "Blob not found in index",
		}
	}
	return
}

// Set attempts to set this blob.
func (b Blob) Set(r io.Reader, ct ContentType) (err error) {
	i.mut.Lock()
	defer i.mut.Unlock()
	file, err := Filepath(b, ct)
	if err != nil {
		return
	}

	bucket, _, err := b.Split()
	if err != nil {
		return
	}

	err = fsio.EnsureDir(fsio.Join(i.Root, bucket))
	if err != nil {
		err = er.Unexpected{Msg: "failed to ensure directory: " + err.Error()}
		return
	}

	i.Set(b, ct)

	err = fsio.WriteAll(file, r)
	if err != nil {
		return
	}

	return
}

// Get attempts to get this blob.
func (b Blob) Get() (r io.Reader, ct ContentType, err error) {
	i.mut.RLock()
	defer i.mut.RUnlock()
	file, ct, err := b.File()
	if err != nil {
		return
	}

	bar, err := os.ReadFile(file)
	if err != nil {
		err = er.NotFound{
			Requested: "Blob",
			From:      "File",
			With:      string(b),
			Msg:       "Blob not found in file",
		}
		return
	}

	r = bytes.NewBuffer(bar)
	return
}

// Del attempts to remove this blob.
func (b Blob) Del() (err error) {
	i.mut.Lock()
	defer i.mut.Unlock()
	file, _, err := b.File()
	if err != nil {
		err = er.NotFound{
			Requested: "Blob",
			From:      "Index",
			With:      string(b),
			Msg:       "Blob not found in index",
		}
		return
	}

	err = os.Remove(file)
	if err != nil {
		return
	}

	if !i.Del(b) {
		err = er.Unexpected{Msg: "blob removed from index before file"}
		return
	}

	return
}
