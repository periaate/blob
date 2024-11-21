package blob

type ContentType int

func (c ContentType) ToString() string {
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

const (
	STREAM ContentType = iota
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

func GetCT(str string) ContentType {
	switch str {
	case "application/octet-stream", "0":
		return STREAM
	case "text/plain", "1":
		return PLAIN
	case "text/html", "2":
		return HTML
	case "application/json", "3":
		return JSON
	case "text/css", "4":
		return CSS
	case "text/javascript", "5":
		return JAVASCRIPT
	case "audio/mp3", "6":
		return MP3
	case "audio/ogg", "7":
		return OGG
	case "image/jpeg", "8":
		return JPEG
	case "image/png", "9":
		return PNG
	case "image/gif", "10":
		return GIF
	case "video/mp4", "11":
		return MP4
	case "video/webm", "12":
		return WEBM
	case "video/mkv", "13":
		return MKV
	default:
		return STREAM
	}
}
