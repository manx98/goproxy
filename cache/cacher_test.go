package cache

import (
	"context"
	"errors"
	"github.com/goproxy/goproxy/utils"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDirCacher(t *testing.T) {
	dirCacher := DirCacher(t.TempDir())

	if rc, err := dirCacher.Get(context.Background(), "a/b/c"); err == nil {
		t.Fatal("expected error")
	} else if got, want := err, fs.ErrNotExist; !utils.CompareErrors(got, want) {
		t.Fatalf("got %q, want %q", got, want)
	} else if got := rc; got != nil {
		t.Errorf("got %#v, want nil", got)
	}

	if err := dirCacher.Put(context.Background(), "a/b/c", strings.NewReader("foobar")); err != nil {
		t.Fatalf("unexpected error %q", err)
	}

	if fi, err := os.Stat(filepath.Join(string(dirCacher), filepath.FromSlash("a/b"))); err != nil {
		t.Fatalf("unexpected error %q", err)
	} else if got, want := fi.Mode().Perm(), os.FileMode(0o755).Perm(); got != want {
		t.Errorf("got %d, want %d", got, want)
	}

	if fi, err := os.Stat(filepath.Join(string(dirCacher), filepath.FromSlash("a/b/c"))); err != nil {
		t.Fatalf("unexpected error %q", err)
	} else if got, want := fi.Mode().Perm(), os.FileMode(0o644).Perm(); got != want {
		t.Errorf("got %d, want %d", got, want)
	}

	if rc, err := dirCacher.Get(context.Background(), "a/b/c"); err != nil {
		t.Fatalf("unexpected error %q", err)
	} else if rc == nil {
		t.Fatal("unexpected nil")
	} else if b, err := io.ReadAll(rc); err != nil {
		t.Fatalf("unexpected error %q", err)
	} else if err := rc.Close(); err != nil {
		t.Fatalf("unexpected error %q", err)
	} else if got, want := string(b), "foobar"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	if err := dirCacher.Put(context.Background(), "d/e/f", &utils.TestReadSeeker{
		ReadSeeker: strings.NewReader("foobar"),
		ReadF: func(rs io.ReadSeeker, p []byte) (n int, err error) {
			return 0, errors.New("cannot read")
		},
	}); err == nil {
		t.Fatal("expected error")
	} else if got, want := err.Error(), "cannot read"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	dirCacher = DirCacher(filepath.Join(string(dirCacher), filepath.FromSlash("a/b/c")))
	if err := dirCacher.Put(context.Background(), "d/e/f", strings.NewReader("foobar")); err == nil {
		t.Fatal("expected error")
	}
}
