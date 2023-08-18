package debug

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"syscall"
	"time"
)

const PanicExitCode = 5

var LogDir string

func SaveLog(after func(p any)) {
	var p = recover()
	if p == nil {
		return
	}

	if after != nil {
		defer after(p)
	}

	var ts = time.Now().Format("20060102030405")
	var filename = "crash." + ts + ".log"
	var path = filepath.Join(LogDir, filename)

	file, err := os.Create(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to dump crash log: %v\n", err)
		fmt.Fprintf(os.Stderr, "panic: %v\n\n", p)
		os.Stderr.Write(debug.Stack())
		return
	}
	defer file.Close()

	fmt.Fprintf(file, "panic: %v\n\n", p)
	file.Write(debug.Stack())

	fmt.Printf("crash log saved to %s\n", path)

}

func Exit(p any) {
	os.Exit(PanicExitCode)
}

func SigInt(p any) {
	syscall.Kill(os.Getpid(), syscall.SIGINT)
	go func() {
		time.Sleep(3 * time.Second) // wait for clean exit
		syscall.Kill(os.Getpid(), syscall.SIGKILL)
	}()
}
