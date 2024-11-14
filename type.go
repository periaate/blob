package blob

import "io"

type Storager interface {
	Add(bType CType, bPath string, r io.Reader) (err error)
	Del(bPath string) (err error)
	Get(bPath string) (r io.ReadCloser, err error)

	Mkdir(dPath string) (err error)
	RmDir(dPath string) (err error)
	Lsdir(dPath string) (blobs [][2]string, err error)
}
