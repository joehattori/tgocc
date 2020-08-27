package main

type (
	ty interface {
		alignment() int
		size() int
	}

	ptr interface {
		ty
		base() ty
	}

	tyArr struct {
		of  ty
		len int
	}

	tyBool  struct{}
	tyChar  struct{}
	tyEmpty struct{}

	tyFn struct {
		retTy ty
	}

	tyInt  struct{}
	tyLong struct{}

	tyPtr struct {
		to ty
	}

	tyShort struct{}

	tyStruct struct {
		align   int
		members []*member
		sz      int
	}

	tyVoid struct{}
)

func newTyArr(of ty, len int) *tyArr { return &tyArr{of, len} }
func newTyBool() *tyBool             { return &tyBool{} }
func newTyChar() *tyChar             { return &tyChar{} }
func newTyEmpty() *tyEmpty           { return &tyEmpty{} }
func newTyFn(t ty) *tyFn             { return &tyFn{t} }
func newTyInt() *tyInt               { return &tyInt{} }
func newTyLong() *tyLong             { return &tyLong{} }
func newTyPtr(to ty) *tyPtr          { return &tyPtr{to} }
func newTyShort() *tyShort           { return &tyShort{} }
func newTyStruct(align int, m []*member, size int) *tyStruct {
	return &tyStruct{align, m, size}
}
func newTyVoid() *tyVoid { return &tyVoid{} }

func alignTo(n int, align int) int {
	return (n + align - 1) / align * align
}

func (a *tyArr) alignment() int    { return a.of.alignment() }
func (b *tyBool) alignment() int   { return 1 }
func (c *tyChar) alignment() int   { return 1 }
func (e *tyEmpty) alignment() int  { return 0 }
func (f *tyFn) alignment() int     { return 1 }
func (i *tyInt) alignment() int    { return 4 }
func (l *tyLong) alignment() int   { return 8 }
func (p *tyPtr) alignment() int    { return 8 }
func (s *tyShort) alignment() int  { return 2 }
func (s *tyStruct) alignment() int { return s.align }
func (v *tyVoid) alignment() int   { return 1 }

func (a *tyArr) size() int    { return a.len * a.of.size() }
func (b *tyBool) size() int   { return 1 }
func (c *tyChar) size() int   { return 1 }
func (e *tyEmpty) size() int  { return 0 }
func (f *tyFn) size() int     { return 1 }
func (i *tyInt) size() int    { return 4 }
func (l *tyLong) size() int   { return 8 }
func (p *tyPtr) size() int    { return 8 }
func (s *tyShort) size() int  { return 2 }
func (s *tyStruct) size() int { return s.sz }
func (v *tyVoid) size() int   { return 1 }

func (a *tyArr) base() ty { return a.of }
func (p *tyPtr) base() ty { return p.to }

func (s *tyStruct) findMember(name string) *member {
	for _, member := range s.members {
		if name == member.name {
			return member
		}
	}
	return nil
}
