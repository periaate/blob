package wrap

import (
	"bytes"
	"io"
	"testing"

	"github.com/periaate/blob"
	"github.com/periaate/blume/clog"
)

func TestWrapper(t *testing.T) {
	clog.SetLogLoggerLevel(clog.LevelDebug)

	wrapper, err := New("http://127.0.0.1:8085", "")
	if err != nil {
		t.Fatalf("failed to create a new wrapper storage %s", err)
	}

	if wrapper == nil {
		t.Fatalf("failed to create a new wrapper storage %s", err)
	}

	defer func() {
		wrapper.Del("test/AAAAAAAAAAAAAAAAAAA")
		wrapper.RmDir("test/")
	}()

	wrapper.Del("test/AAAAAAAAAAAAAAAAAAA")
	wrapper.RmDir("test/")

	if err := wrapper.MkDir("test/"); err != nil {
		t.Fatalf("failed to create a new directory %s", err)
	}

	if err := wrapper.MkDir("test/"); err == nil {
		t.Fatalf("created a directory that already exists %s", err)
	}

	if err := wrapper.Add(blob.STREAM, "test/AAAAAAAAAAAAAAAAAAA", bytes.NewBuffer([]byte("test"))); err != nil {
		t.Fatalf("failed to add a blob %s", err)
	}

	if err := wrapper.Add(blob.STREAM, "test/AAAAAAAAAAAAAAAAAAA", bytes.NewBuffer([]byte("test"))); err == nil {
		t.Fatalf("added a blob that already exists %s", err)
	}

	if val, _, err := wrapper.Get("test/AAAAAAAAAAAAAAAAAAA"); err != nil {
		t.Fatalf("failed to get a blob %s", err)
	} else {
		defer val.Close()
		bar := bytes.NewBuffer([]byte{})
		io.Copy(bar, val)
		if bar.String() != "test" {
			t.Fatalf("got an unexpected blob %s", err)
		}
	}

	if err := wrapper.Del("test/AAAAAAAAAAAAAAAAAAA"); err != nil {
		t.Fatalf("failed to delete a blob %s", err)
	}

	if err := wrapper.Del("test/AAAAAAAAAAAAAAAAAAA"); err == nil {
		t.Fatalf("deleted a blob that does not exist %s", err)
	}

	if err := wrapper.RmDir("test/"); err != nil {
		t.Fatalf("failed to delete a directory %s", err)
	}
}
