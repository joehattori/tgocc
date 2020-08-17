package main

// TyKind is a type variable
type TyKind int

// Type represents type
type Type interface {
	kind() TyKind
}

const (
	// TyKindInt is a type of int
	TyKindInt = iota
	// TyKindPtr is a type of pointer
	TyKindPtr
)

// TyInt represents int type
type TyInt struct{}

func (*TyInt) kind() TyKind {
	return TyKindInt
}

// TyPtr represents pointer type
type TyPtr struct {
	to Type
}

func (*TyPtr) kind() TyKind {
	return TyKindPtr
}
