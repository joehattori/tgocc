package types

type (
	// Type is the interface of types.
	Type interface {
		Alignment() int
		Size() int
	}

	// Pointing is the interface of pointer and array type.
	Pointing interface {
		Type
		Base() Type
	}

	// Arr represents array type.
	Arr struct {
		Of  Type
		Len int
	}

	Bool  struct{}
	Char  struct{}
	Empty struct{}
	Enum  struct{}
	Fn    struct {
		RetTy Type
	}

	Int  struct{}
	Long struct{}

	Ptr struct {
		To Type
	}

	Short struct{}

	Struct struct {
		Align   int
		Members []*Member
		Sz      int
	}

	Void struct{}
)

func NewArr(of Type, len int) *Arr { return &Arr{of, len} }
func NewBool() *Bool               { return &Bool{} }
func NewChar() *Char               { return &Char{} }
func NewEmpty() *Empty             { return &Empty{} }
func NewEnum() *Enum               { return &Enum{} }
func NewFn(t Type) *Fn             { return &Fn{t} }
func NewInt() *Int                 { return &Int{} }
func NewLong() *Long               { return &Long{} }
func NewPtr(to Type) *Ptr          { return &Ptr{to} }
func NewShort() *Short             { return &Short{} }
func NewStruct(align int, m []*Member, Size int) *Struct {
	return &Struct{align, m, Size}
}
func NewVoid() *Void { return &Void{} }

func NewMember(name string, offset int, t Type) *Member {
	return &Member{name, offset, t}
}

func AlignTo(n int, align int) int {
	return (n + align - 1) / align * align
}

func (a *Arr) Alignment() int    { return a.Of.Alignment() }
func (b *Bool) Alignment() int   { return 1 }
func (c *Char) Alignment() int   { return 1 }
func (e *Empty) Alignment() int  { return 0 }
func (e *Enum) Alignment() int   { return 4 }
func (f *Fn) Alignment() int     { return 1 }
func (i *Int) Alignment() int    { return 4 }
func (l *Long) Alignment() int   { return 8 }
func (p *Ptr) Alignment() int    { return 8 }
func (s *Short) Alignment() int  { return 2 }
func (s *Struct) Alignment() int { return s.Align }
func (v *Void) Alignment() int   { return 1 }

func (a *Arr) Size() int    { return a.Len * a.Of.Size() }
func (b *Bool) Size() int   { return 1 }
func (c *Char) Size() int   { return 1 }
func (e *Empty) Size() int  { return 0 }
func (e *Enum) Size() int   { return 4 }
func (f *Fn) Size() int     { return 1 }
func (i *Int) Size() int    { return 4 }
func (l *Long) Size() int   { return 8 }
func (p *Ptr) Size() int    { return 8 }
func (s *Short) Size() int  { return 2 }
func (s *Struct) Size() int { return s.Sz }
func (v *Void) Size() int   { return 1 }

func (a *Arr) Base() Type { return a.Of }
func (p *Ptr) Base() Type { return p.To }

type Member struct {
	Name   string
	Offset int
	Type   Type
}

// FindMember retrieves the struct member with the given name.
func (s *Struct) FindMember(name string) *Member {
	for _, member := range s.Members {
		if name == member.Name {
			return member
		}
	}
	return nil
}
