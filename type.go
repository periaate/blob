package blob

import "io"

var _ Storager = &Storage{}

type Storager interface {
	Add(bType CType, bPath string, r io.Reader) (err error)
	Del(bPath string) (err error)
	Get(bPath string) (r io.ReadCloser, ctype CType, err error)

	MkDir(dPath string) (err error)
	RmDir(dPath string) (err error)
	LsDir(dPath string) (blobs [][2]string, err error)
}

type CType int

func (c CType) String() string {
	switch c {
	case STREAM:
		return "application/octet-stream"
	case PLAIN:
		return "text/plain"
	case HTML:
		return "text/html"
	case JSON:
		return "application/json"
	case CSS:
		return "text/css"
	case JAVASCRIPT:
		return "text/javascript"
	case MP3:
		return "audio/mp3"
	case OGG:
		return "audio/ogg"
	case JPEG:
		return "image/jpeg"
	case PNG:
		return "image/png"
	case GIF:
		return "image/gif"
	case MP4:
		return "video/mp4"
	case WEBM:
		return "video/webm"
	case MKV:
		return "video/mkv"
	default:
		return "application/octet-stream"
	}
}

func ContentType(str string) CType {
	switch str {
	case "text/plain":
		return PLAIN
	case "application/octet-stream":
		return STREAM
	case "text/html":
		return HTML
	case "application/json":
		return JSON
	case "text/css":
		return CSS
	case "text/javascript":
		return JAVASCRIPT
	case "audio/mp3":
		return MP3
	case "audio/ogg":
		return OGG
	case "image/jpeg":
		return JPEG
	case "image/png":
		return PNG
	case "image/gif":
		return GIF
	case "video/mp4":
		return MP4
	case "video/webm":
		return WEBM
	case "video/mkv":
		return MKV
	default:
		return STREAM
	}
}

const (
	STREAM CType = iota
	PLAIN
	HTML
	JSON
	CSS
	JAVASCRIPT
	MP3
	OGG
	JPEG
	PNG
	GIF
	MP4
	WEBM
	MKV
)
