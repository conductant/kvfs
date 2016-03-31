package kvfs

import (
	"bazil.org/fuse/fs"
)

type FS struct {
	db *Backend
}

var _ = fs.FS(&FS{})

func (f *FS) Root() (fs.Node, error) {
	n := &Dir{
		fs: f,
	}
	return n, nil
}
