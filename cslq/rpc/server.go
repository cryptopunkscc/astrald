package rpc

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/cslq"
	"io"
	"net"
	"reflect"
	"time"
)

const (
	ResponseOK = iota
	ResponseError
)

type Server struct {
	exports map[string]*funcWrapper
	Log     io.Writer
}

func ServerFromStruct(v interface{}) *Server {
	srv := &Server{}
	if err := srv.ExportStruct(v); err != nil {
		panic(err)
	}
	return srv
}

// Serve accepts connections from the listener and serves them
func (srv *Server) Serve(listener net.Listener) error {
	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}

		if srv.Log != nil {
			fmt.Fprintf(srv.Log, "[%s] connected (%s).\n", conn.RemoteAddr(), conn.RemoteAddr().Network())
		}

		go func() {
			srv.serveConn(conn)
			fmt.Fprintf(srv.Log, "[%s] disconnected.\n", conn.RemoteAddr())
		}()
	}
}

// ServeCall serves a single call over the provided transport. On error, CallInfo may contain partial information.
func (srv *Server) ServeCall(rw io.ReadWriter) (info CallInfo, err error) {
	var argBytes []byte

	if err = cslq.Decode(rw, "[c]c [l]c", &info.Name, &argBytes); err != nil {
		return
	}

	fn, ok := srv.exports[info.Name]
	if !ok {
		info.Response = ResponseError
		err = cslq.Encode(rw, "c", info.Response)
		return
	}

	var (
		argReader = bytes.NewReader(argBytes)
		argFmt    = fn.argFormat()
		argVals   = fn.argTemplate()
		argPtrs   = valuesToPointers(argVals)
	)

	// Inject a Context if possible
	if len(argPtrs) > 0 {
		if _, ok := argPtrs[0].(*context.Context); ok {
			argPtrs, argFmt = argPtrs[1:], argFmt[1:]

			var ctx = context.Background()
			if conn, ok := rw.(net.Conn); ok {
				ctx = context.WithValue(ctx, "RemoteEndpoint", conn.RemoteAddr())
			}
			argVals[0] = reflect.ValueOf(ctx)
		}
	}

	// Decode arguments
	if err = argFmt.Decode(argReader, argPtrs...); err != nil {
		return
	}

	info.Args = valuesToInterfaces(argVals)

	// Call the function
	info.CallStart = time.Now()
	retVals := fn.Call(argVals)
	info.CallEnd = time.Now()

	info.Vals = valuesToInterfaces(retVals)

	// if last value is an error, convert it to a pointer
	if len(retVals) > 0 {
		var errType = reflect.TypeOf((*error)(nil)).Elem()
		lastVal := retVals[len(retVals)-1]
		if lastVal.Type().ConvertibleTo(errType) {
			var newErr = errors.New("")
			if !lastVal.IsNil() {
				var oldErr = lastVal.Interface().(error)
				newErr = errors.New(oldErr.Error())
			}
			retVals[len(retVals)-1] = reflect.ValueOf(&newErr)
		}
	}

	if err = cslq.Encode(rw, "c", ResponseOK); err != nil {
		return
	}

	var retBuffer = &bytes.Buffer{}

	if err = fn.retFormat().Encode(retBuffer, valuesToInterfaces(retVals)...); err != nil {
		return
	}

	err = cslq.Encode(rw, "[l]c", retBuffer.Bytes())

	return
}

func (srv *Server) ExportFunc(f interface{}, prefixArgs ...interface{}) error {
	return srv.ExportFuncAs(f, "", prefixArgs...)
}

func (srv *Server) ExportFuncAs(f interface{}, exportName string, prefixArgs ...interface{}) error {
	if srv.exports == nil {
		srv.exports = make(map[string]*funcWrapper)
	}

	var rv = reflect.ValueOf(f)

	for rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	if rv.Kind() != reflect.Func {
		return errors.New("not a function")
	}

	if exportName == "" {
		exportName = rv.Type().Name()
	}

	if exportName == "" {
		return errors.New("cannot export anonymous function")
	}

	srv.exports[exportName] = wrapFuncWithPrefixArgs(rv, prefixArgs...)

	return nil
}

func (srv *Server) ExportStruct(v interface{}) error {
	var rv = reflect.ValueOf(v)

	if !isStructOrPtr(rv) {
		return errors.New("not a struct")
	}

	for i := 0; i < rv.NumMethod(); i++ {
		if !rv.Type().Method(i).IsExported() {
			continue
		}

		var method = rv.Type().Method(i)
		if err := srv.ExportFuncAs(method.Func.Interface(), method.Name, v); err != nil {
			return err
		}
	}

	return nil
}

func (srv *Server) serveConn(conn net.Conn) error {
	for {
		info, err := srv.ServeCall(conn)
		if err != nil {
			return err
		}
		if srv.Log != nil {
			if info.Response == ResponseOK {
				fmt.Fprintf(srv.Log, "[%s] executed %s in %v\n", conn.RemoteAddr(), info.Name, info.Duration())
			} else {
				fmt.Fprintf(srv.Log, "[%s] invalid method: %s\n", conn.RemoteAddr(), info.Name)
			}
		}
	}
}
