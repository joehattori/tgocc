package main

// Var represents a variable
type Var interface {
	getType() Type
}

// LVar represents a local variable
type LVar struct {
	name   string
	ty     Type
	offset int
}

// GVar represents a global variable
type GVar struct {
	name string
	ty   Type
}

func (v *LVar) getType() Type {
	return v.ty
}

func (v *GVar) getType() Type {
	return v.ty
}

// NewLVar creates a new instance of LVar
func NewLVar(name string, ty Type, offset int) *LVar {
	return &LVar{name, ty, offset}
}

// NewGVar creates a new instance of GVar
func NewGVar(name string, ty Type) *GVar {
	return &GVar{name, ty}
}
