package main

import (
	"flag"
	"fmt"
	astral2 "github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/astral"
	"github.com/cryptopunkscc/astrald/streams"
	"io"
	"net"
	"os"
	"os/exec"
	"strings"
	"sync"
)

var wait = true

const (
	exitSuccess = iota
	exitHelp
	exitError
)

func log(f string, v ...any) {
	fmt.Fprintf(os.Stderr, f, v...)
	if !strings.HasSuffix(f, "\n") {
		fmt.Fprintf(os.Stderr, "\n")
	}
}

func cmdExport(args []string) {
	if len(args) < 2 {
		log("anc export <astral_port> <local_address>")
		os.Exit(exitHelp)
	}

	var serviceName = args[0]
	var dstAddr = args[1]

	service, err := astral.Register(serviceName)
	if err != nil {
		log("register error: %s", err)
		os.Exit(exitError)
	}

	log("forwarding %s to %s", serviceName, dstAddr)

	for query := range service.QueryCh() {
		var peerName = displayName(query.RemoteIdentity())

		log("[%s] connected.", peerName)

		dstConn, err := net.Dial("tcp", dstAddr)
		if err != nil {
			log("[%s] error dialing %s: %s\n", peerName, dstAddr, err)
			query.Reject()
			continue
		}

		srcConn, _ := query.Accept()
		go func() {
			wc, rc, _ := streams.Join(srcConn, dstConn)
			log("[%s] disconnected (%d read, %d written).", peerName, rc, wc)
		}()
	}

}

func cmdImport(args []string) {
	if len(args) < 3 {
		log("anc import <local_address> <astral_node> <astral_port>")
		os.Exit(exitHelp)
	}

	srcAddr := args[0]
	nodeName := args[1]
	serviceName := args[2]

	srcListener, err := net.Listen("tcp", srcAddr)
	if err != nil {
		log("listen error: %s", err)
		os.Exit(exitError)
	}

	log("forwarding %s to %s:%s\n", srcAddr, nodeName, serviceName)
	i := 0

	for {
		srcConn, err := srcListener.Accept()
		if err != nil {
			log("accept error: %s", err)
			continue
		}
		var conn = (srcConn).(*astral.Conn)
		var peerName = displayName(conn.RemoteIdentity())

		log("[%s] connected.", peerName)

		dstConn, err := astral.QueryName(nodeName, serviceName)
		if err != nil {
			log("[%s] query %s:%s error: %s", peerName, nodeName, serviceName, err)
			conn.Close()
			continue
		}

		go func(i int) {
			wc, rc, _ := streams.Join(conn, dstConn)
			log("[%s] disconnected (%d read, %d written).", peerName, rc, wc)
		}(i)
	}
}

func cmdRegister(args []string) {
	if len(args) < 1 {
		log("anc register <name>")
		os.Exit(exitHelp)
	}

	serviceName := args[0]

	l, err := astral.Register(serviceName)
	if err != nil {
		log("register error: %s", err.Error())
		os.Exit(exitError)
	}

	log("listening on %s", serviceName)

	nconn, err := l.Accept()
	if err != nil {
		log("accept error: %s", err)
		return
	}

	var conn = nconn.(*astral.Conn)
	var peerName = displayName(conn.RemoteIdentity())

	log("[%s] connected.", peerName)

	go func() {
		n, _ := io.Copy(conn, os.Stdin)
		if !wait {
			conn.Close()
		}
		log("[%s] wrote %d bytes", peerName, n)
	}()

	n, _ := io.Copy(os.Stdout, conn)
	log("[%s] read %d bytes", peerName, n)

	os.Exit(exitSuccess)
}

func cmdShare(args []string) {
	if len(args) < 2 {
		log("anc share <name> <file>")
		os.Exit(exitHelp)
	}

	var serviceName = args[0]
	var filename = args[1]

	if _, err := os.Stat(filename); err != nil {
		log("file not found: %s", filename)
		os.Exit(exitError)
	}

	port, err := astral.Register(serviceName)
	if err != nil {
		log("register error: %s", err.Error())
		os.Exit(exitError)
	}

	log("serving %s on %s", filename, serviceName)

	for conn := range port.AcceptAll() {
		var conn = conn.(*astral.Conn)
		var peerName = displayName(conn.RemoteIdentity())

		log("[%s] connected.\n", peerName)

		file, err := os.Open(filename)
		if err != nil {
			log("[%s] error opening file: %s\n", peerName, err)
			os.Exit(exitError)
		}

		go func() {
			defer conn.Close()
			n, err := io.Copy(conn, file)
			if err != nil {
				log("[%s] write error: %s\n", peerName, err)
				return
			}
			log("[%s] sent %d bytes\n", peerName, n)
		}()
	}
}

func cmdExec(args []string) {
	if len(args) < 2 {
		log("anc exec <name> <exec_path> [args]")
		os.Exit(exitHelp)
	}

	serviceName := args[0]
	execPath := args[1]
	args = args[2:]

	service, err := astral.Register(serviceName)
	if err != nil {
		log("register error: %s", err)
		os.Exit(exitError)
	}

	log("listening on %s", serviceName)

	for conn := range service.AcceptAll() {
		conn := conn.(*astral.Conn)

		var peerName = displayName(conn.RemoteIdentity())

		log("[%s] connected.", peerName)

		go func() {
			proc := exec.Command(execPath, args...)
			stdin, _ := proc.StdinPipe()
			stdout, _ := proc.StdoutPipe()
			stderr, _ := proc.StderrPipe()
			err = proc.Start()
			if err != nil {
				log("exec error: %s", err)
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

			log("[%s] disconnected.", peerName)
		}()
	}

	os.Exit(exitSuccess)
}

func cmdQuery(args []string) {
	var nodeID, query string

	if len(args) < 1 {
		log("anc query [node] <query>")
		os.Exit(exitHelp)
	}

	query = args[0]
	if len(args) > 1 {
		nodeID = args[0]
		query = args[1]
	}

	conn, err := astral.QueryName(nodeID, query)
	if err != nil {
		log("error: %s", err)
		os.Exit(exitError)
	}
	var peerName = displayName(conn.RemoteIdentity())

	log("connected.")

	go func() {
		n, _ := io.Copy(conn, os.Stdin)
		if !wait {
			conn.Close()
		}
		log("[%s] wrote %d bytes", peerName, n)
	}()

	n, _ := io.Copy(os.Stdout, conn)
	log("[%s] read %d bytes", peerName, n)

	os.Exit(exitSuccess)
}

func cmdDownload(args []string) {
	var nodeID, query, filename string

	if len(args) < 2 {
		log("anc download <node> <service> [filename]")
		os.Exit(exitHelp)
	}

	nodeID = args[0]
	query = args[1]
	filename = query
	if len(args) >= 3 {
		filename = args[2]
	}

	if _, err := os.Stat(filename); err == nil {
		log("file already exists: %s", filename)
		os.Exit(exitError)
	}

	conn, err := astral.QueryName(nodeID, query)
	if err != nil {
		log("error: %s", err)
		os.Exit(exitError)
	}

	log("connected.")

	file, err := os.Create(filename)
	if err != nil {
		log("error creating file %s: %s", filename, err)
		os.Exit(exitError)
	}

	n, err := io.Copy(file, conn)
	if err != nil {
		log("download error: %s", err)
		os.Exit(exitError)
	}

	log("read %d bytes", n)
	os.Exit(exitSuccess)
}

func cmdResolve(args []string) {
	if len(args) < 1 {
		log("anc resolve <name>")
		os.Exit(exitHelp)
	}

	identity, err := astral.Resolve(args[0])
	if err != nil {
		log("error:", err)
		os.Exit(exitError)
	}
	fmt.Println(identity.String())

	nodeInfo, err := astral.GetNodeInfo(identity)
	if err != nil {
		return
	}

	fmt.Println("alias", nodeInfo.Name)
}

func help() {
	log("astral netcat")
	log("usage: anc <query|register|exec|share|download|resolve|help>")
	os.Exit(exitHelp)
}

func main() {
	if len(os.Args) < 2 {
		help()
	}

	flag.BoolVar(&wait, "w", false, "wait for remote EOF")
	flag.Parse()

	var args = flag.Args()

	cmd := args[0]
	switch cmd {
	case "q", "query":
		cmdQuery(args[1:])
	case "r", "register":
		cmdRegister(args[1:])
	case "e", "exec":
		cmdExec(args[1:])
	case "s", "share":
		cmdShare(args[1:])
	case "d", "download":
		cmdDownload(args[1:])
	case "resolve":
		cmdResolve(args[1:])
	case "export":
		cmdExport(args[1:])
	case "import":
		cmdImport(args[1:])
	case "h", "help":
		help()
	default:
		help()
	}
}

func displayName(identity *astral2.Identity) string {
	if info, err := astral.GetNodeInfo(identity); err == nil {
		return info.Name
	}
	return identity.String()
}
