package blob

import "fmt"

func getKind(isDir bool) string {
	if isDir {
		return "directory"
	}
	return "file"
}

type ErrExists struct {
	Path  string
	IsDir bool
}

func (e ErrExists) Error() string {
	return fmt.Sprintf("%s exists: %s", getKind(e.IsDir), e.Path)
}

func ErrIsExists(err error) bool {
	_, ok := err.(ErrExists)
	return ok
}

type ErrNoDir struct {
	Path string
}

func (e ErrNoDir) Error() string {
	return fmt.Sprintf("no directory: %s", e.Path)
}

func ErrIsNoDir(err error) bool {
	_, ok := err.(ErrNoDir)
	return ok
}

type ErrNotExists struct {
	Path  string
	IsDir bool
}

func (e ErrNotExists) Error() string {
	return fmt.Sprintf("%s not exists: %s", getKind(e.IsDir), e.Path)
}

func ErrIsNotExists(err error) bool {
	_, ok := err.(ErrNotExists)
	return ok
}

type ErrIsDir struct {
	Path string
}

func (e ErrIsDir) Error() string {
	return fmt.Sprintf("is directory: %s", e.Path)
}

func ErrIsIsDir(err error) bool {
	_, ok := err.(ErrIsDir)
	return ok
}

type ErrBadPath struct {
	Path  string
	IsDir bool
}

func (e ErrBadPath) Error() string {
	return fmt.Sprintf("bad path: %s", e.Path)
}

func ErrIsBadPath(err error) bool {
	_, ok := err.(ErrBadPath)
	return ok
}

type ErrUnsupportedContentType struct {
	Path string
	Type CType
}

func (e ErrUnsupportedContentType) Error() string {
	return fmt.Sprintf("unsupported content type: %s for %s", e.Type.String(), e.Path)
}

func ErrIsUnsupportedContentType(err error) bool {
	_, ok := err.(ErrUnsupportedContentType)
	return ok
}
