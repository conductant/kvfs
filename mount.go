package kvfs

import (
	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"io"
	"os"
)

type handle struct {
	io.Closer

	conn *fuse.Conn
}

func (this *handle) Close() error {
	if this.conn != nil {
		return this.conn.Close()
	}
	return nil
}

// Mount does not block.  It's up to the caller to block by reading on a channel, etc.
func Mount(url, mountpoint string, config *Config) (io.Closer, error) {
	db, err := NewBackend(url, config)
	if err != nil {
		return nil, err
	}

	var perm os.FileMode = 0644

	if err := os.MkdirAll(mountpoint, perm); err != nil {
		return nil, err
	}

	c, err := fuse.Mount(mountpoint)
	if err != nil {
		return nil, err
	}

	go func() {
		fs.Serve(c, &FS{
			db: db,
		})
	}()
	return &handle{conn: c}, nil
}

func Unmount(mountpoint string) error {
	return fuse.Unmount(mountpoint)
}
