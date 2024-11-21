package blob

import (
	"bytes"
	"os"
	"testing"

	"github.com/periaate/blume/clog"
)

func TestStorage(t *testing.T) {
	clog.SetLogLoggerLevel(clog.LevelDebug)

	st, err := NewStorage("./testing")
	if st == nil {
		clog.Error("NewStorage returned a nil value")
		t.FailNow()
	}

	if err != nil {
		clog.Error("NewStorage returned an error", "err", err)
		t.FailNow()
	}

	err = st.Set("bucket/blob", PLAIN, bytes.NewBufferString("hello"))
	if err != nil {
		clog.Error("set returned an error", "err", err)
		t.FailNow()
	}

	fp, err := st.Get("bucket/blob")
	if err != nil {
		clog.Error("Get returned an error", "err", err)
		t.FailNow()
	}

	bar, err := os.ReadFile(fp)
	if err != nil {
		clog.Error("ReadFile returned an error", "err", err)
		t.FailNow()
	}

	if string(bar) != "hello" {
		clog.Error("ReadFile returned an unexpected value", "bar", bar)
		t.FailNow()
	}
}
