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

	// s := blob.New("test")
	// if s == nil {
	// 	t.Fatal("failed to create a new blob storage")
	// }
	//
	// go func() {
	// 	addr := "127.0.0.1:8085"
	// 	clog.Info("starting server", "addr", fmt.Sprintf("http://%s", addr))
	// 	log.Fatal(http.ListenAndServe(addr, s))
	// }()
	//

	wrapper, err := New("http://127.0.0.1:8085", "")
	if err != nil {
		t.Fatalf("failed to create a new wrapper storage %s", err)
	}

	if wrapper == nil {
		t.Fatalf("failed to create a new wrapper storage %s", err)
	}

	wrapper.Del("test/blob")
	wrapper.RmDir("test")

	if err := wrapper.Mkdir("test"); err != nil {
		t.Fatalf("failed to create a new directory %s", err)
	}

	if err := wrapper.Mkdir("test"); err == nil {
		t.Fatalf("created a directory that already exists %s", err)
	}

	if err := wrapper.Add(blob.STREAM, "test/blob", bytes.NewBuffer([]byte("test"))); err != nil {
		t.Fatalf("failed to add a blob %s", err)
	}

	if err := wrapper.Add(blob.STREAM, "test/blob", bytes.NewBuffer([]byte("test"))); err == nil {
		t.Fatalf("added a blob that already exists %s", err)
	}

	if val, err := wrapper.Get("test/blob"); err != nil {
		t.Fatalf("failed to get a blob %s", err)
	} else {
		defer val.Close()
		bar := bytes.NewBuffer([]byte{})
		io.Copy(bar, val)
		if bar.String() != "test" {
			t.Fatalf("got an unexpected blob %s", err)
		}
	}

	if err := wrapper.Del("test/blob"); err != nil {
		t.Fatalf("failed to delete a blob %s", err)
	}

	if err := wrapper.Del("test/blob"); err == nil {
		t.Fatalf("deleted a blob that does not exist %s", err)
	}

	if err := wrapper.RmDir("test"); err != nil {
		t.Fatalf("failed to delete a directory %s", err)
	}
}