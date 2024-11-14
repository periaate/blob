package blob

import "testing"

func TestErr(t *testing.T) {
	a := ErrExists{Path: "a", IsDir: true}
	b := ErrNotExists{Path: "b", IsDir: false}
	c := ErrBadPath{Path: "c", IsDir: true}
	d := ErrNoDir{Path: "d"}
	e := ErrIsDir{Path: "e"}
	f := ErrDirNotEmpty{Path: "f"}
	g := ErrIllegalPath{Path: "g"}

	if !ErrIs[ErrExists](a) {
		t.Error("ErrExists")
	}
	if !ErrIs[ErrNotExists](b) {
		t.Error("ErrNotExists")
	}
	if !ErrIs[ErrBadPath](c) {
		t.Error("ErrBadPath")
	}
	if !ErrIs[ErrNoDir](d) {
		t.Error("ErrNoDir")
	}
	if !ErrIs[ErrIsDir](e) {
		t.Error("ErrIsDir")
	}
	if !ErrIs[ErrDirNotEmpty](f) {
		t.Error("ErrDirNotEmpty")
	}
	if !ErrIs[ErrIllegalPath](g) {
		t.Error("ErrIllegalPath")
	}

	if ErrIs[ErrExists](g) {
		t.Error("ErrExists")
	}
	if ErrIs[ErrNotExists](f) {
		t.Error("ErrNotExists")
	}
	if ErrIs[ErrBadPath](e) {
		t.Error("ErrBadPath")
	}
	if ErrIs[ErrIsDir](c) {
		t.Error("ErrIsDir")
	}
	if ErrIs[ErrDirNotEmpty](b) {
		t.Error("ErrDirNotEmpty")
	}
	if ErrIs[ErrIllegalPath](a) {
		t.Error("ErrIllegalPath")
	}
}
