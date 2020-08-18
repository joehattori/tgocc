package main

// Type represents type
type Type interface{}

// TyInt represents int type
type TyInt struct{}

// TyPtr represents pointer type
type TyPtr struct {
	to Type
}
