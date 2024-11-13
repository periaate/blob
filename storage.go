package blob

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/periaate/blume/clog"
	"github.com/periaate/blume/fsio"
	"github.com/periaate/blume/gen"
	"github.com/periaate/blume/str"
)

type Storage struct {
	Root string
}

/*
blob
	add: add new blob
	set: set existing blob
	del: delete existing blob
	get: get existing blob
*/

func getPath(root string, bPath string) (fp string, exists bool, isDir bool, err error) {
	if str.Contains("..")(bPath) {
		return "", false, false, ErrIllegalPath{bPath}
	}
	fp = fsio.Join(root, bPath)
	exists = fsio.Exists(fp)
	isDir = str.HasSuffix("/")(bPath)
	return
}

func (s *Storage) Add(bType CType, bPath string, r io.Reader) (err error) {
	fp, exists, isDir, err := getPath(s.Root, bPath)
	switch {
	case ErrIsIllegalPath(err):
		return
	case exists:
		err = ErrExists{Path: bPath}
	case isDir:
		err = ErrIsDir{Path: bPath}
	case !fsio.Exists(fsio.Dir(fp)):
		err = ErrNoDir{Path: fsio.Dir(bPath)}
	}

	if err != nil {
		return
	}

	base := fsio.Base(fp)
	dir := fsio.Dir(fp)
	fp = fsio.Join(dir, fmt.Sprintf("%d_%s", bType, base))

	err = fsio.WriteNew(fp, r)
	if os.IsNotExist(err) {
		err = ErrNoDir{Path: bPath}
	}
	return
}

func (s *Storage) Get(bPath string) (rc io.ReadCloser, cType CType, err error) {
	fp, _, isDir, err := getPath(s.Root, bPath)
	switch {
	case ErrIsIllegalPath(err):
		return
	case isDir:
		err = ErrIsDir{Path: bPath}
	default:
		fp, err = findBlob(s.Root, bPath)
		if err != nil {
			return
		}

		blob := [2]string{}

		blob, err = SplitBlob(fsio.Base(fp))
		if err != nil {
			return
		}

		cType = ContentType(blob[0])
		rc, err = fsio.Open(fp)
	}

	return
}

func findBlob(root string, bPath string) (fp string, err error) {
	fp = fsio.Join(root, bPath)
	files, err := fsio.ReadDir(fsio.Dir(fp))
	if err != nil {
		err = ErrNoDir{Path: fsio.Dir(bPath)}
		return
	}

	fp = gen.First(str.HasSuffix(fsio.Base(bPath)))(files)
	if fp == "" {
		err = ErrNotExists{Path: bPath}
	}

	return
}

func (s *Storage) Del(bPath string) (err error) {
	fp, _, isDir, err := getPath(s.Root, bPath)
	switch {
	case ErrIsIllegalPath(err):
		return
	case isDir:
		err = ErrIsDir{Path: bPath}
	default:
		fp, err = findBlob(s.Root, bPath)
		if err != nil {
			return
		}
		err = fsio.Remove(fp)
	}

	return
}

/*
tree
	mkdir: make new directory
	rmdir: delete existing directory
	lsdir: list blobs in directory
*/

func (s *Storage) Mkdir(dPath string) (err error) {
	fp, exists, isDir, err := getPath(s.Root, dPath)
	switch {
	case ErrIsIllegalPath(err):
		return
	case exists:
		err = ErrExists{Path: dPath, IsDir: isDir}
	case !isDir:
		err = ErrBadPath{Path: dPath}
	default:
		err = fsio.EnsureDir(fp)
	}

	return
}

func (s *Storage) Rmdir(dPath string) (err error) {
	fp, exists, isDir, err := getPath(s.Root, dPath)
	switch {
	case ErrIsIllegalPath(err):
		return
	case !exists:
		err = ErrNotExists{Path: dPath, IsDir: isDir}
	case !isDir:
		err = ErrBadPath{Path: dPath}
	}

	res, err := fsio.ReadDir(fp)
	if err != nil {
		return
	}

	if len(res) > 0 {
		err = ErrDirNotEmpty{Path: dPath}
	} else {
		err = fsio.Remove(fp)
	}

	return
}

func (s *Storage) Lsdir(dPath string) (blobs [][2]string, err error) {
	var res []string
	fp, exists, isDir, err := getPath(s.Root, dPath)
	switch {
	case ErrIsIllegalPath(err):
		return
	case !exists:
		err = ErrNotExists{Path: dPath}
	case !isDir:
		err = ErrBadPath{Path: dPath, IsDir: isDir}
	default:
		res, err = fsio.ReadDir(fp)
	}

	if err != nil {
		return
	}

	var blob [2]string
	for _, p := range res {
		blob, err = SplitBlob(fsio.Base(p))
		if err != nil {
			return
		}

		blobs = append(blobs, blob)
	}

	return
}

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

	contentType := CType(enumValue).String()

	res[0] = contentType
	res[1] = parts[1]
	clog.Debug("split blob", "type", contentType, "name", parts[1])
	return
}
