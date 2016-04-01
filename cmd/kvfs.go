package main

import (
	"fmt"
	"github.com/conductant/gohm/pkg/command"
	"github.com/conductant/gohm/pkg/runtime"
	"github.com/conductant/kvfs"
	"io"
)

func main() {

	config := &struct {
		kvfs.Config

		MountPath string `flag:"m,Mount path"`
		Url       string `flag:"url,Url to backend"`
	}{}

	command.RegisterFunc("mount", config,
		func(a []string, w io.Writer) error {
			url := config.Url
			if url == "" {
				if len(a) < 1 {
					return fmt.Errorf("No url specified")
				} else {
					url = a[0]
				}
			}
			mountPath := config.MountPath
			if mountPath == "" {
				if len(a) < 2 {
					return fmt.Errorf("No mount point specified.")
				} else {
					mountPath = a[1]
				}
			}

			return kvfs.Mount(url, mountPath, &config.Config)
		},
		func(w io.Writer) {
			fmt.Fprintln(w, "Mount backend by url to local file system path")
		})

	runtime.Main()
}
