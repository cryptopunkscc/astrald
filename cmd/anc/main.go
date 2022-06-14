package main

import (
	"fmt"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/lib/astral"
	"github.com/cryptopunkscc/astrald/logfmt"
	"github.com/cryptopunkscc/astrald/streams"
	"io"
	"net"
	"os"
	"os/exec"
	"strings"
	"sync"
)

const (
	exitSuccess = iota
	exitHelp
	exitError
)

func cmdExport(args []string) {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "anc export <astral_port> <local_address>")
		os.Exit(exitHelp)
	}

	portName := args[0]
	dstAddr := args[1]

	port, err := astral.Listen(portName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "register error: %s\n", err.Error())
		os.Exit(exitError)
	}

	fmt.Fprintf(os.Stderr, "exporting %s to %s\n", portName, dstAddr)

	i := 0

	for query := range port.QueryCh() {
		i += 1
		fmt.Printf("[%d] %s connected.\n", i, displayName(query.RemoteIdentity()))

		dstConn, err := net.Dial("tcp", dstAddr)
		if err != nil {
			fmt.Printf("error: destination (%s) unreachable: %v\n", dstAddr, err)
			query.Reject()
			continue
		}

		srcConn, _ := query.Accept()
		go func(i int) {
			streams.Join(srcConn, dstConn)
			fmt.Printf("[%d] %s disconnected.\n", i, displayName(query.RemoteIdentity()))
		}(i)
	}

}

func cmdImport(args []string) {
	if len(args) < 3 {
		fmt.Fprintln(os.Stderr, "anc import <local_address> <astral_node> <astral_port>")
		os.Exit(exitHelp)
	}

	srcAddr := args[0]
	nodeName := args[1]
	portName := args[2]

	srcListener, err := net.Listen("tcp", srcAddr)
	if err != nil {
		fmt.Println("listen error:", err)
		os.Exit(exitError)
	}

	fmt.Fprintf(os.Stderr, "importing %s to %s:%s\n", srcAddr, nodeName, portName)
	i := 0

	for {
		srcConn, err := srcListener.Accept()
		if err != nil {
			fmt.Println("accept error:", err)
			continue
		}

		i += 1
		fmt.Printf("[%d] %s connected.\n", i, srcConn.RemoteAddr())

		dstConn, err := astral.DialName(nodeName, portName)
		if err != nil {
			fmt.Println("astral.dial error:", err)
			srcConn.Close()
			continue
		}

		go func(i int) {
			streams.Join(srcConn, dstConn)
			fmt.Printf("[%d] %s disconnected.\n", i, srcConn.RemoteAddr())
		}(i)
	}
}

func cmdRegister(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "anc register <name>")
		os.Exit(exitHelp)
	}

	portName := args[0]

	l, err := astral.Listen(portName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "register error: %s\n", err.Error())
		os.Exit(exitError)
	}

	fmt.Fprintf(os.Stderr, "listening on %s\n", portName)

	conn, err := l.Accept()
	if err != nil {
		return
	}

	fmt.Fprintf(os.Stderr, "%s connected.\n", addrName(conn.RemoteAddr()))

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

	port, err := astral.Listen(portName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "register error: %s\n", err.Error())
		os.Exit(exitError)
	}

	fmt.Fprintf(os.Stderr, "listening on %s\n", portName)

	for conn := range port.AcceptAll() {
		fmt.Fprintln(os.Stderr, addrName(conn.RemoteAddr()), "connected.")

		file, err := os.Open(filename)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error opening file:", err)
			os.Exit(exitError)
		}

		conn := conn
		go func() {
			defer conn.Close()
			n, err := io.Copy(conn, file)
			if err != nil {
				fmt.Fprintln(os.Stderr, conn.RemoteAddr(), "error sending file:", err)
				return
			}
			fmt.Fprintln(os.Stderr, addrName(conn.RemoteAddr()), "downloaded", logfmt.DataSize(n).HumanReadable())
		}()
	}
}

func cmdExec(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "anc exec <name> <exec_path>")
		os.Exit(exitHelp)
	}

	portName := args[0]
	execPath := args[1]

	l, err := astral.Listen(portName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: register: %s\n", err.Error())
		os.Exit(exitError)
	}

	fmt.Fprintf(os.Stderr, "listening on %s\n", portName)

	for conn := range l.AcceptAll() {
		conn := conn

		fmt.Fprintf(os.Stderr, "%s connected.\n", addrName(conn.RemoteAddr()))

		go func() {
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

			fmt.Fprintln(os.Stderr, conn.RemoteAddr(), "disconnected.")
		}()
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

	conn, err := astral.DialName(nodeID, query)
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

	conn, err := astral.DialName(nodeID, query)
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

func cmdResolve(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "anc resolve <name>")
		os.Exit(exitHelp)
	}

	identity, err := astral.Resolve(args[0])
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(exitError)
	}
	fmt.Println(identity.String())

	nodeInfo, err := astral.NodeInfo(identity)
	if err != nil {
		return
	}

	fmt.Println("name", nodeInfo.Name)
}

func help() {
	fmt.Fprintln(os.Stderr, "astral netcat")
	fmt.Fprintln(os.Stderr, "usage: anc <query|register|exec|share|download|resolve|help>")
	os.Exit(exitHelp)
}

func main() {
	if len(os.Args) < 2 {
		help()
	}

	// Check if ANC_PROTO is set in the environment
	for _, env := range os.Environ() {
		env = strings.ToLower(env)
		parts := strings.SplitN(env, "=", 2)
		if parts[0] == "anc_proto" {
			astral.ListenProtocol = parts[1]
			break
		}
	}

	cmd := os.Args[1]
	switch cmd {
	case "q", "query":
		cmdQuery(os.Args[2:])
	case "r", "register":
		cmdRegister(os.Args[2:])
	case "e", "exec":
		cmdExec(os.Args[2:])
	case "s", "share":
		cmdShare(os.Args[2:])
	case "d", "download":
		cmdDownload(os.Args[2:])
	case "resolve":
		cmdResolve(os.Args[2:])
	case "export":
		cmdExport(os.Args[2:])
	case "import":
		cmdImport(os.Args[2:])
	case "h", "help":
		help()
	default:
		help()
	}
}

func displayName(identity id.Identity) string {
	if info, err := astral.NodeInfo(identity); err == nil {
		return info.Name
	}
	return identity.String()
}

func addrName(addr net.Addr) string {
	if addr.Network() != "astral" {
		return addr.String()
	}

	identity, err := id.ParsePublicKeyHex(addr.String())
	if err != nil {
		return addr.String()
	}

	return displayName(identity)
}
