package blob

import (
	"fmt"

	"github.com/periaate/blume/gen"
)

type ErrExists struct {
	Path  string
	IsDir bool
}

func (e ErrExists) Error() string {
	return fmt.Sprintf("%s exists: %s", gen.Tern(e.IsDir, "dir", "file"), e.Path)
}

type ErrNoDir struct{ Path string }

func (e ErrNoDir) Error() string { return fmt.Sprintf("no directory: %s", e.Path) }

type ErrNotExists struct {
	Path  string
	IsDir bool
}

func (e ErrNotExists) Error() string {
	return fmt.Sprintf("%s not exists: %s", gen.Tern(e.IsDir, "dir", "file"), e.Path)
}

type ErrIsDir struct{ Path string }

func (e ErrIsDir) Error() string { return fmt.Sprintf("is directory: %s", e.Path) }

type ErrBadPath struct {
	Path  string
	IsDir bool
}

func (e ErrBadPath) Error() string { return fmt.Sprintf("bad path: %s", e.Path) }

type ErrUnsupportedContentType struct {
	Path string
	Type ContentType
}

func (e ErrUnsupportedContentType) Error() string {
	return fmt.Sprintf("unsupported content type: %s for %s", e.Type.String(), e.Path)
}

type ErrBadRequest struct{ Msg string }

func (e ErrBadRequest) Error() string { return fmt.Sprintf("bad request: %s", e.Msg) }

type ErrFatal struct{ Msg string }

func (e ErrFatal) Error() string { return fmt.Sprintf("fatal error: %s", e.Msg) }

type ErrDirNotEmpty struct{ Path string }

func (e ErrDirNotEmpty) Error() string { return fmt.Sprintf("directory %s is not empty", e.Path) }

type ErrIllegalPath struct{ Path string }

func (e ErrIllegalPath) Error() string { return fmt.Sprintf("illegal path: %s", e.Path) }

func ErrIs[A any](err error) bool {
	_, ok := err.(A)
	return ok
}

func StatusOfErr(err error) int {
	switch err.(type) {
	case ErrIllegalPath:
		return 400
	case ErrExists:
		return 409
	case ErrNoDir:
		return 404
	case ErrNotExists:
		return 404
	case ErrIsDir:
		return 409
	case ErrBadRequest:
		return 400
	case ErrDirNotEmpty:
		return 400
	case ErrBadPath:
		return 400
	case ErrUnsupportedContentType:
		return 415
	case ErrFatal:
		return 500
	default:
		return 500
	}
}
