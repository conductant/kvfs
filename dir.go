package kvfs

import (
	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"errors"
	"golang.org/x/net/context"
	"os"
)

type Dir struct {
	fs *FS
	// path from root to this dir; empty for root dir
	path []string
}

var _ = fs.Node(&Dir{})

func (d *Dir) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Mode = os.ModeDir | 0755
	return nil
}

var _ = fs.HandleReadDirAller(&Dir{})

func (d *Dir) ReadDirAll(c context.Context) ([]fuse.Dirent, error) {
	var res []fuse.Dirent
	err := d.fs.db.View(c, func(ctx Context) error {
		b := ctx.Dir(d.path)
		if b == nil {
			return errors.New("dir no longer exists")
		}
		for entry := range b.Cursor() {
			de := fuse.Dirent{
				Name: entry.Key,
			}
			if entry.Dir {
				de.Type = fuse.DT_Dir
			} else {
				de.Type = fuse.DT_File
			}
			res = append(res, de)
		}
		return nil
	})
	return res, err
}

var _ = fs.NodeStringLookuper(&Dir{})

func (d *Dir) Lookup(c context.Context, name string) (fs.Node, error) {
	var n fs.Node
	err := d.fs.db.View(c, func(ctx Context) error {
		b := ctx.Dir(d.path)
		if b == nil {
			return errors.New("dir no longer exists")
		}
		if child := b.Dir(name); child != nil {
			// directory
			dirs := make([]string, len(d.path)+1)
			dirs = append(dirs, d.path...)
			dirs = append(dirs, name)
			n = &Dir{
				fs:   d.fs,
				path: dirs,
			}
			return nil
		}
		if child := b.Get(name); child != nil {
			// file
			n = &File{
				dir:  d,
				name: name,
			}
			return nil
		}
		return fuse.ENOENT
	})
	if err != nil {
		return nil, err
	}
	return n, nil
}

var _ = fs.NodeMkdirer(&Dir{})

func (d *Dir) Mkdir(c context.Context, req *fuse.MkdirRequest) (fs.Node, error) {
	name := req.Name
	err := d.fs.db.Update(c, func(ctx Context) error {
		b := ctx.Dir(d.path)
		if b == nil {
			return errors.New("dir no longer exists")
		}
		if child := b.Dir(name); child != nil {
			return fuse.EEXIST
		}
		if _, err := b.CreateDir(name); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	dirs := make([]string, len(d.path)+1)
	dirs = append(dirs, d.path...)
	dirs = append(dirs, name)
	n := &Dir{
		fs:   d.fs,
		path: dirs,
	}
	return n, nil
}

var _ = fs.NodeCreater(&Dir{})

func (d *Dir) Create(ctx context.Context, req *fuse.CreateRequest, resp *fuse.CreateResponse) (fs.Node, fs.Handle, error) {

	name := req.Name
	f := &File{
		dir:     d,
		name:    name,
		writers: 1,
		// file is empty at Create time, no need to set data
	}
	return f, f, nil
}

var _ = fs.NodeRemover(&Dir{})

func (d *Dir) Remove(c context.Context, req *fuse.RemoveRequest) error {
	name := req.Name
	return d.fs.db.Update(c, func(ctx Context) error {
		b := ctx.Dir(d.path)
		if b == nil {
			return errors.New("dir no longer exists")
		}

		switch req.Dir {
		case true:
			if b.Dir(name) == nil {
				return fuse.ENOENT
			}
			if err := b.DeleteDir(name); err != nil {
				return err
			}

		case false:
			if b.Get(name) == nil {
				return fuse.ENOENT
			}
			if err := b.Delete(name); err != nil {
				return err
			}
		}
		return nil
	})
}
