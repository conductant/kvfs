package main

import (
	"fmt"
	"github.com/conductant/gohm/pkg/command"
	"github.com/conductant/gohm/pkg/runtime"
	"github.com/conductant/kvfs"
	"io"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	fromKernel := make(chan os.Signal)

	// kill -9 is SIGKILL and is uncatchable.
	signal.Notify(fromKernel, syscall.SIGHUP)  // 1
	signal.Notify(fromKernel, syscall.SIGINT)  // 2
	signal.Notify(fromKernel, syscall.SIGQUIT) // 3
	signal.Notify(fromKernel, syscall.SIGABRT) // 6
	signal.Notify(fromKernel, syscall.SIGTERM) // 15

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

			closer, err := kvfs.Mount(url, mountPath, &config.Config)
			if err != nil {
				return err
			}

			for range fromKernel {
				fmt.Println("Unmounting", mountPath)
				if err := kvfs.Unmount(mountPath); err == nil {
					return nil
				} else {
					fmt.Println("Cannot unmount. Err=", err)
				}
			}
			return closer.Close()
		},
		func(w io.Writer) {
			fmt.Fprintln(w, "Mount backend by url to local file system path.")
			fmt.Fprintln(w, "Usage: kvfs mount <flags> | <url> <mountpoint>")
		})

	runtime.Main()

}
