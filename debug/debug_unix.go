//go:build !windows

package debug

import (
	"os"
	"syscall"
	"time"
)

func SigInt(p any) {
	syscall.Kill(os.Getpid(), syscall.SIGINT)
	go func() {
		time.Sleep(3 * time.Second) // wait for clean exit
		syscall.Kill(os.Getpid(), syscall.SIGKILL)
	}()
}
