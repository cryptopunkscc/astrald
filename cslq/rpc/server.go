package rpc

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/cslq"
	"io"
	"net"
	"reflect"
)

const (
	ResponseOK = iota
	ResponseError
)

type Server struct {
	exports map[string]*funcWrapper
}

func ServerFromStruct(v interface{}) *Server {
	srv := &Server{}
	if err := srv.ExportStruct(v); err != nil {
		panic(err)
	}
	return srv
}

func Serve(listener net.Listener, server *Server) error {
	return server.Serve(listener)
}

func ServeStruct(listener net.Listener, v interface{}) error {
	return Serve(listener, ServerFromStruct(v))
}

func (srv *Server) Serve(listener net.Listener) error {
	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		srv.handleClient(conn)
	}
	return nil
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

func (srv *Server) handleClient(rw io.ReadWriter) error {
	for {
		if err := srv.handleCall(rw); err != nil {
			return err
		}
	}
}

func (srv *Server) handleCall(rw io.ReadWriter) error {
	var funcName string

	if err := cslq.Decode(rw, "[c]c", &funcName); err != nil {
		return err
	}

	fn, ok := srv.exports[funcName]
	if !ok {
		return cslq.Encode(rw, "c", ResponseError)
	}

	// Decode arguments
	var (
		argFmt  = fn.argFormat()
		argVals = fn.argTemplate()
		argPtrs = valuesToPointers(argVals)
	)

	if len(argPtrs) > 0 {
		if _, ok := argPtrs[0].(*context.Context); ok {
			argPtrs = argPtrs[1:]
			argFmt = argFmt[1:]

			ctx := context.Background()

			if conn, ok := rw.(net.Conn); ok {
				ctx = context.WithValue(ctx, "RemoteAddr", conn.RemoteAddr())
			}

			argVals[0] = reflect.ValueOf(ctx)
		}
	}

	if err := argFmt.Decode(rw, argPtrs...); err != nil {
		return err
	}

	// Call the function
	retVals := fn.Call(argVals)

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

	if err := cslq.Encode(rw, "c", ResponseOK); err != nil {
		return err
	}

	return fn.retFormat().Encode(rw, valuesToInterfaces(retVals)...)
}
