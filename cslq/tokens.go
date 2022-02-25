package cslq

type Token interface{}

type ArrayStartToken struct{}
type ArrayEndToken struct{}
type Uint8Token struct{}
type Uint16Token struct{}
type Uint32Token struct{}
type Uint64Token struct{}
type NumberLiteralToken uint64
type StructStartToken struct{}
type StructEndToken struct{}
type InterfaceToken struct{}
