package apphost

import (
	"github.com/cryptopunkscc/astrald/astral"
	"io"
)

// This file contains all data structures and values used by the apphost protocol.

// Return codes

const (
	Success = iota
	Rejected
	AlreadyRegistered
)

// TokenArgs contains arguments for the token method
type TokenArgs struct {
	Token astral.String8
}

var _ astral.Object = &TokenArgs{}

func (TokenArgs) ObjectType() string { return "astrald.mod.apphost.token_args" }

func (t TokenArgs) WriteTo(w io.Writer) (n int64, err error) {
	return astral.Struct(t).WriteTo(w)
}

func (t *TokenArgs) ReadFrom(r io.Reader) (n int64, err error) {
	return astral.Struct(t).ReadFrom(r)
}

// TokenResponse contains response values for the token method
type TokenResponse struct {
	Code    astral.Uint8
	GuestID *astral.Identity
	HostID  *astral.Identity
}

var _ astral.Object = &TokenResponse{}

func (TokenResponse) ObjectType() string { return "astrald.mod.apphost.token_response" }

func (t TokenResponse) WriteTo(w io.Writer) (n int64, err error) {
	return astral.Struct(t).WriteTo(w)
}

func (t *TokenResponse) ReadFrom(r io.Reader) (n int64, err error) {
	return astral.Struct(t).ReadFrom(r)
}

// RegisterArgs contains arguments for the register method
type RegisterArgs struct {
	Identity *astral.Identity
	Endpoint astral.String8
	Flags    astral.Uint8
}

var _ astral.Object = &RegisterArgs{}

func (RegisterArgs) ObjectType() string {
	return "astrald.mod.apphost.register_args"
}

func (args RegisterArgs) WriteTo(w io.Writer) (n int64, err error) {
	return astral.Struct(args).WriteTo(w)
}

func (args *RegisterArgs) ReadFrom(r io.Reader) (n int64, err error) {
	return astral.Struct(args).ReadFrom(r)
}

// RegisterResponse contains response values for the register method
type RegisterResponse struct {
	Code  astral.Uint8
	Token astral.String8
}

var _ astral.Object = &RegisterResponse{}

func (res RegisterResponse) ObjectType() string {
	return "astrald.mod.apphost.register_response"
}

func (res RegisterResponse) WriteTo(w io.Writer) (n int64, err error) {
	return astral.Struct(res).WriteTo(w)
}

func (res *RegisterResponse) ReadFrom(r io.Reader) (n int64, err error) {
	return astral.Struct(res).ReadFrom(r)
}

// QueryArgs contains arguments for the query method
type QueryArgs struct {
	Caller *astral.Identity
	Target *astral.Identity
	Query  astral.String16
}

var _ astral.Object = &QueryArgs{}

func (QueryArgs) ObjectType() string { return "astrald.mod.apphost.query_args" }

func (q QueryArgs) WriteTo(w io.Writer) (n int64, err error) {
	return astral.Struct(q).WriteTo(w)
}

func (q *QueryArgs) ReadFrom(r io.Reader) (n int64, err error) {
	return astral.Struct(q).ReadFrom(r)
}

// QueryResponse contains response values for the query method
type QueryResponse struct {
	Code astral.Uint8
}

var _ astral.Object = &QueryResponse{}

func (QueryResponse) ObjectType() string { return "astrald.mod.apphost.query_response" }

func (q QueryResponse) WriteTo(w io.Writer) (n int64, err error) {
	return astral.Struct(q).WriteTo(w)
}

func (q *QueryResponse) ReadFrom(r io.Reader) (n int64, err error) {
	return astral.Struct(q).ReadFrom(r)
}

// QueryInfo contains information about an en route query
type QueryInfo struct {
	Token  astral.String8
	Caller *astral.Identity
	Target *astral.Identity
	Query  astral.String16
}

var _ astral.Object = &QueryInfo{}

func (QueryInfo) ObjectType() string { return "astrald.mod.apphost.query_info" }

func (q QueryInfo) WriteTo(w io.Writer) (n int64, err error) {
	return astral.Struct(q).WriteTo(w)
}

func (q *QueryInfo) ReadFrom(r io.Reader) (n int64, err error) {
	return astral.Struct(q).ReadFrom(r)
}
