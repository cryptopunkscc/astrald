package main

import (
	"fmt"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/logfmt"
	"github.com/cryptopunkscc/astrald/mod/apphost/client"
	"io"
	"os"
	"os/exec"
	"sync"
)

const (
	exitSuccess = iota
	exitHelp
	exitError
)

func cmdRegister(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "anc register <name>")
		os.Exit(exitHelp)
	}

	portName := args[0]

	port, err := astral.Reqister(portName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "register error: %s\n", err.Error())
		os.Exit(exitError)
	}

	fmt.Fprintf(os.Stderr, "listening on %s\n", portName)

	req := <-port.Next()
	if req == nil {
		return
	}

	fmt.Fprintf(os.Stderr, "%s connected.\n", req.Caller())

	conn, err := req.Accept()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err.Error())
		os.Exit(exitError)
	}

	go func() {
		io.Copy(conn, os.Stdin)
		conn.Close()
	}()

	io.Copy(os.Stdout, conn)
	os.Exit(exitSuccess)
}

func cmdShare(args []string) {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "anc share <name> <file>")
		os.Exit(exitHelp)
	}

	portName := args[0]
	filename := args[1]

	if _, err := os.Stat(filename); err != nil {
		fmt.Fprintf(os.Stderr, "file not found: %s\n", filename)
		os.Exit(exitError)
	}

	port, err := astral.Reqister(portName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "register error: %s\n", err.Error())
		os.Exit(exitError)
	}

	fmt.Fprintf(os.Stderr, "listening on %s\n", portName)

	for req := range port.Next() {
		caller, _ := id.ParsePublicKeyHex(req.Caller())

		fmt.Fprintln(os.Stderr, logfmt.ID(caller), "connected.")
		conn, _ := req.Accept()

		file, err := os.Open(filename)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error opening file:", err)
			os.Exit(exitError)
		}

		go func(caller id.Identity) {
			defer conn.Close()
			n, err := io.Copy(conn, file)
			if err != nil {
				fmt.Fprintln(os.Stderr, logfmt.ID(caller), "error sending file:", err)
				return
			}
			fmt.Fprintln(os.Stderr, logfmt.ID(caller), "downloaded", logfmt.DataSize(n).HumanReadable())
		}(caller)
	}
}

func cmdExec(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "anc exec <name> <exec_path>")
		os.Exit(exitHelp)
	}

	portName := args[0]
	execPath := args[1]

	port, err := astral.Reqister(portName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: register: %s\n", err.Error())
		os.Exit(exitError)
	}

	fmt.Fprintf(os.Stderr, "listening on %s\n", portName)

	for req := range port.Next() {
		conn, err := req.Accept()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %s\n", err.Error())
			os.Exit(exitError)
		}

		fmt.Fprintf(os.Stderr, "%s connected.\n", req.Caller())

		go func(req *astral.Request) {
			proc := exec.Command(execPath)
			stdin, _ := proc.StdinPipe()
			stdout, _ := proc.StdoutPipe()
			stderr, _ := proc.StderrPipe()
			err = proc.Start()
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(exitError)
			}

			var wg sync.WaitGroup
			wg.Add(3)

			// Stdout
			go func() {
				io.Copy(conn, stdout)
				conn.Close()
				wg.Done()
			}()

			// Stderr
			go func() {
				io.Copy(conn, stderr)
				conn.Close()
				wg.Done()
			}()

			// Stdin
			go func() {
				io.Copy(stdin, conn)
				proc.Process.Kill()
				wg.Done()
			}()

			wg.Wait()
			proc.Wait()

			fmt.Fprintln(os.Stderr, req.Caller(), "disconnected.")
		}(req)
	}
	os.Exit(exitSuccess)
}

func cmdQuery(args []string) {
	var nodeID, query string

	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "anc query [nodeid] <query>")
		os.Exit(exitHelp)
	}

	query = args[0]
	if len(args) > 1 {
		nodeID = args[0]
		query = args[1]
	}

	conn, err := astral.Query(nodeID, query)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(exitError)
	}

	fmt.Fprintf(os.Stderr, "connected.\n")

	go func() {
		io.Copy(conn, os.Stdin)
		conn.Close()
	}()

	io.Copy(os.Stdout, conn)
	os.Exit(exitSuccess)
}

func cmdDownload(args []string) {
	var nodeID, query, filename string

	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "anc download <nodeid> <query> [filename]")
		os.Exit(exitHelp)
	}

	nodeID = args[0]
	query = args[1]
	filename = query
	if len(args) >= 3 {
		filename = args[2]
	}

	if _, err := os.Stat(filename); err == nil {
		fmt.Fprintf(os.Stderr, "file already exists: %s\n", filename)
		os.Exit(exitError)
	}

	conn, err := astral.Query(nodeID, query)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(exitError)
	}

	fmt.Fprintf(os.Stderr, "connected.\n")

	file, err := os.Create(filename)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error creating file:", err)
		os.Exit(exitError)
	}

	n, err := io.Copy(file, conn)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(exitError)
	}

	fmt.Fprintln(os.Stderr, "Downloaded", logfmt.DataSize(n).HumanReadable())

	io.Copy(os.Stdout, conn)
	os.Exit(exitSuccess)
}

func help() {
	fmt.Fprintln(os.Stderr, "astral netcat")
	fmt.Fprintln(os.Stderr, "usage: anc <query|register|exec|share|download|help>")
	os.Exit(exitHelp)
}

func main() {
	if len(os.Args) < 2 {
		help()
	}

	cmd := os.Args[1]
	switch cmd[0] {
	case 'q':
		cmdQuery(os.Args[2:])
	case 'r':
		cmdRegister(os.Args[2:])
	case 'e':
		cmdExec(os.Args[2:])
	case 's':
		cmdShare(os.Args[2:])
	case 'd':
		cmdDownload(os.Args[2:])
	case 'h':
		help()
	default:
		help()
	}
}
