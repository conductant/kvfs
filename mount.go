package kvfs

import (
	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

func Mount(url, mountpoint string, config *Config) error {
	db, err := NewBackend(url, config)
	if err != nil {
		return err
	}

	c, err := fuse.Mount(mountpoint)
	if err != nil {
		return err
	}
	defer c.Close()

	filesys := &FS{
		db: db,
	}
	if err := fs.Serve(c, filesys); err != nil {
		return err
	}

	// check if the mount process has an error to report
	<-c.Ready
	if err := c.MountError; err != nil {
		return err
	}

	return nil
}
