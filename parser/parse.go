package parser

import (
	"log"

	"github.com/joehattori/tgocc/ast"
	"github.com/joehattori/tgocc/tokenizer"
	"github.com/joehattori/tgocc/types"
	"github.com/joehattori/tgocc/vars"
)

// Parser holds the structure defining a parser object.
type Parser struct {
	curFnName string
	curScope  *scope
	Ast       *ast.Ast
	Toks      []tokenizer.Token
}

// NewParser creates a new parser.
func NewParser(toks []tokenizer.Token) *Parser {
	return &Parser{"", &scope{}, &ast.Ast{}, toks}
}

func (p *Parser) findVar(s string) vars.Var {
	v := p.searchVar(s)
	if v == nil {
		log.Fatalf("Undefined vars.Var %s", s)
	}
	return v
}

func (p *Parser) searchStructTag(tag string) *structTag {
	scope := p.curScope
	for scope != nil {
		if tag := scope.searchStructTag(tag); tag != nil {
			return tag
		}
		scope = scope.super
	}
	return nil
}

func (p *Parser) searchEnumTag(tag string) *enumTag {
	scope := p.curScope
	for scope != nil {
		if tag := scope.searchEnumTag(tag); tag != nil {
			return tag
		}
		scope = scope.super
	}
	return nil
}

func (p *Parser) searchVar(varName string) vars.Var {
	scope := p.curScope
	for scope != nil {
		if v := scope.searchVar(varName); v != nil {
			return v
		}
		scope = scope.super
	}
	return nil
}

/*
Actual parsing process from here.

	program    = (function | globalVar)*
	function   = baseType types.TypeDecl "(" (ident ("," ident)* )? ")" ("{" stmt* "}" | ";")
	globalVar  = decl
	stmt       = expr ";"
  				| "{" stmt* "}"
  				| "return" expr ";"
  				| "if" "(" expr ")" stmt ("else" stmt) ?
  				| "while" "(" expr ")" stmt
				| "do" stmt "while" "(" expr ")" ";"
  				| "for" "(" (expr? ";" | decl) expr? ";" expr? ")" stmt
				| "typedef" types.Type ident ("[" constExpr "]")* ";"
				| "switch" "(" expr ")" "{" switchCase* ("default" ":" stmt*)? }"
  				| decl
	switchCase = "case" num ":" stmt*
	decl       = baseType types.TypeDecl ("[" constExpr "]")* "=" initialize ;" | baseTy ";"
	tyDecl     = "*"* (ident | "(" types.TypeDecl ")")
	expr       = assign
	assign     = ternary (("=" | "+=" | "-=" | "*=" | "/=") assign) ?
	ternary    = logOr ("?" expr ":" ternary)?
	logOr      = logAnd ("||" logAnd)*
	logAnd     = bitOr ("&&" bitOr)*
	bitOr      = bitXor ("|" bitXor)*
	bitXor     = bitAnd ("^" bitAnd)*
	bitAnd     = equality ("&" equality)*
	equality   = relational ("==" relational | "!=" relational)*
	relational = shift ("<" shift | "<=" shift | ">" shift | ">=" shift)*
	shift      = add ("<<" add | ">>" add | "<<=" add | ">>=" add)
	add        = mul ("+" mul | "-" mul)*
	mul        = cat ("*" cast | "/" cast)*
	cast       = "(" baseType "*"*  ")" cast | unary
	unary      = ("+" | "-" | "*" | "&" | "!" | "~")? cast | ("++" | "--") unary | postfix
	postfix    = primary (("[" expr "]") | ("." ident) | ("->" ident) | "++" | "--")*
	primary    =  num
				| "sizeof" unary
				| str
				| ident ("(" (expr ("," expr)* )? ")")?
				| "(" expr ")"
				| stmtExpr
	stmtExpr   = "(" "{" stmt+ "}" ")"
*/

// Parse traverses tokens and generates Ast.
func (p *Parser) Parse() {
	ast := p.Ast
	for !p.isEOF() {
		if p.isFunction() {
			if fn := p.function(); fn != nil {
				ast.Fns = append(ast.Fns, fn)
			}
		} else {
			if ty, id, rhs, sc := p.decl(); ty != nil {
				init := buildGVarInit(ty, rhs)
				emit := (sc & extern) == 0
				p.curScope.addGVar(emit, id, ty, init)
				ast.GVars = append(ast.GVars, vars.NewGVar(emit, id, ty, init))
			}
		}
	}
}

func buildGVarInit(t types.Type, rhs ast.Node) vars.GVarInit {
	if rhs == nil {
		return nil
	}
	switch t := t.(type) {
	case *types.Arr:
		switch rhs := rhs.(type) {
		case *ast.BlkNode:
			var body []vars.GVarInit
			idx := 0
			for _, e := range rhs.Body {
				idx++
				body = append(body, buildGVarInit(t.Of, e))
			}
			if t.Len < 0 {
				t.Len = idx
			}
			// zero out the rest
			if _, ok := t.Of.(*types.Arr); !ok {
				for i := idx; i < t.Len; i++ {
					body = append(body, vars.NewGVarInitZero(t.Of.Size()))
				}
			}
			return vars.NewGVarInitArr(body)
		default:
			if ok, str := isStrNode(rhs); ok {
				if t.Len < 0 {
					t.Len = len(str)
				} else if t.Len > len(str) {
					for i := len(str); i < t.Len; i++ {
						str += string('\000')
					}
				}
				return vars.NewGVarInitStr(str)
			}
			log.Fatalf("Unhandled case in global vars.Var initialization: %T", rhs)
			return nil
		}
	case *types.Struct:
		var body []vars.GVarInit
		idx := 0
		for i, e := range rhs.(*ast.BlkNode).Body {
			idx++
			mem := t.Members[i]
			if e == nil {
				body = append(body, vars.NewGVarInitZero(mem.Type.Size()))
				continue
			}
			var toAppend []vars.GVarInit
			toAppend = append(toAppend, buildGVarInit(mem.Type, e))
			// padding for struct members
			var end int
			if i < len(t.Members)-1 {
				end = t.Members[i+1].Offset
			} else {
				end = t.Size()
			}
			start := mem.Offset + mem.Type.Size()
			if end-start > 0 {
				toAppend = append(toAppend, vars.NewGVarInitZero(end-start))
			}
			body = append(body, vars.NewGVarInitArr(toAppend))
		}
		return vars.NewGVarInitArr(body)
	default:
		switch rhs := rhs.(type) {
		case *ast.AddrNode:
			return vars.NewGVarInitLabel(rhs.Var.(*ast.VarNode).Var.Name())
		case *ast.VarNode:
			if _, ok := rhs.Var.Type().(*types.Arr); ok {
				return vars.NewGVarInitLabel(rhs.Var.Name())
			}
		}
		return vars.NewGVarInitInt(ast.Eval(rhs), t.Size())
	}
}

func (p *Parser) function() *ast.FnNode {
	ty, _, sc := p.baseType()
	fnName, ty := p.tyDecl(ty)
	p.curFnName = fnName
	p.curScope.addGVar((sc&static) != 0, fnName, types.NewFn(ty), nil)
	fn := ast.NewFnNode((sc&static) != 0, fnName, ty)
	p.spawnScope()
	p.readFnParams(fn)
	if p.consume(";") {
		p.rewindScope()
		return nil
	}
	p.expect("{")
	for !p.consume("}") {
		fn.Body = append(fn.Body, p.stmt())
	}
	p.setFnLVars(fn)
	p.rewindScope()
	// TODO: align
	fn.StackSize = p.curScope.curOffset
	return fn
}

type storageClass int

const (
	static storageClass = 0b01
	extern storageClass = 0b10
)

func (p *Parser) decl() (t types.Type, id string, rhs ast.Node, sc storageClass) {
	t, isTypeDef, sc := p.baseType()
	if p.consume(";") {
		return
	}
	id, t = p.tyDecl(t)
	if isTypeDef {
		p.expect(";")
		p.curScope.addTypeDef(id, t)
		// returned t is nil when it is types.Typepedef (no need to add to scope.vars)
		return nil, "", nil, sc
	}
	t = p.tySuffix(t)
	if p.consume(";") {
		return
	}
	if (sc & extern) != 0 {
		p.expect(";")
		return
	}
	p.expect("=")
	rhs = p.initializer(t, sc)
	p.expect(";")
	return
}

func (p *Parser) initializer(t types.Type, sc storageClass) ast.Node {
	switch t := t.(type) {
	case *types.Arr:
		var nodes []ast.Node
		if strTok, ok := p.consumeStr(); ok {
			init := vars.NewGVarInitStr(strTok.Str())
			s := vars.NewGVar((sc&static) != 0, newGVarLabel(), types.NewArr(types.NewChar(), strTok.Len()), init)
			p.Ast.GVars = append(p.Ast.GVars, s)
			return ast.NewVarNode(s)
		}
		p.expect("{")
		for !p.consume("}") {
			nodes = append(nodes, p.initializer(t.Of, sc))
			if !p.consume(",") {
				p.expect("}")
				break
			}
		}
		return ast.NewBlkNode(nodes)
	case *types.Struct:
		nodes := make([]ast.Node, len(t.Members))
		if strTok, ok := p.consumeStr(); ok {
			init := vars.NewGVarInitStr(strTok.Str())
			s := vars.NewGVar((sc&static) != 0, newGVarLabel(), types.NewArr(types.NewChar(), strTok.Len()), init)
			p.Ast.GVars = append(p.Ast.GVars, s)
			return ast.NewVarNode(s)
		}
		p.expect("{")
		idx := 0
		for !p.consume("}") {
			if p.consume(".") {
				id := p.expectID().Str()
				p.expect("=")
				for i, mem := range t.Members[idx:] {
					if id == mem.Name {
						idx += i
						break
					}
				}
			}
			nodes[idx] = p.initializer(t.Members[idx].Type, sc)
			if !p.consume(",") {
				p.expect("}")
				break
			}
			idx++
		}
		return ast.NewBlkNode(nodes)
	default:
		return p.assign()
	}
}

func (p *Parser) baseType() (t types.Type, isTypeDef bool, sc storageClass) {
	if p.consume("typedef") {
		isTypeDef = true
	}
	if p.consume("static") {
		sc |= static
	}
	if p.consume("extern") {
		sc |= extern
	}
	if isTypeDef && (sc != 0) || (sc == 0b11) {
		log.Fatal("typedef, static and extern should not be used together.")
	}
	switch tok := p.Toks[0].(type) {
	case *tokenizer.IDTok:
		if typeDef, ok := p.searchVar(tok.Str()).(*vars.TypeDef); ok {
			p.popToks()
			return typeDef.Type(), isTypeDef, sc
		}
	case *tokenizer.ReservedTok:
		if p.beginsWith("struct") {
			return p.structDecl(), isTypeDef, sc
		}
		if p.beginsWith("enum") {
			return p.enumDecl(), isTypeDef, sc
		}
		if p.consume("void") {
			return types.NewVoid(), isTypeDef, sc
		}
		if p.consume("_Bool") {
			return types.NewBool(), isTypeDef, sc
		}
		// just skip `volatile` keyword since there is no optimisation yet.
		p.consume("volatile")
		// TODO: handle unsigned properly
		p.consume("unsigned")
		p.consume("signed")
		if p.consume("char") {
			return types.NewChar(), isTypeDef, sc
		}
		if p.consume("short") {
			p.consume("int")
			return types.NewShort(), isTypeDef, sc
		}
		if p.consume("int") {
			return types.NewInt(), isTypeDef, sc
		}
		if p.consume("long") {
			p.consume("long")
			p.consume("int")
			return types.NewLong(), isTypeDef, sc
		}
		return types.NewInt(), isTypeDef, sc
	}
	log.Fatalf("Type expected but got %T: %s", p.Toks[0], p.Toks[0].Str())
	return
}

func (p *Parser) tyDecl(baseTy types.Type) (id string, ty types.Type) {
	for p.consume("*") {
		baseTy = types.NewPtr(baseTy)
	}
	if p.consume("(") {
		id, ty = p.tyDecl(nil)
		p.expect(")")
		baseTy = p.tySuffix(baseTy)
		switch ty := ty.(type) {
		case *types.Arr:
			ty.Of = baseTy
		case *types.Ptr:
			ty.To = baseTy
		default:
			ty = baseTy
		}
		return
	}
	return p.expectID().Str(), p.tySuffix(baseTy)
}

func (p *Parser) tySuffix(t types.Type) types.Type {
	if !p.consume("[") {
		return t
	}
	l := -1
	if !p.consume("]") {
		l = int(p.constExpr())
		p.expect("]")
	}
	t = p.tySuffix(t)
	return types.NewArr(t, l)
}

func (p *Parser) constExpr() int64 {
	return ast.Eval(p.ternary())
}

func (p *Parser) readFnParams(fn *ast.FnNode) {
	p.expect("(")
	isFirstArg := true
	orig := p.Toks
	if p.consume("void") && p.consume(")") {
		return
	}
	p.Toks = orig
	for !p.consume(")") {
		if !isFirstArg {
			p.expect(",")
		}
		isFirstArg = false

		ty, _, _ := p.baseType()
		id, ty := p.tyDecl(ty)
		if arr, ok := ty.(*types.Arr); ok {
			ty = types.NewPtr(arr.Of)
		}
		lv := p.curScope.addLVar(id, ty)
		fn.Params = append(fn.Params, lv)
	}
}

func (p *Parser) setFnLVars(fn *ast.FnNode) {
	offset := 0
	for _, sv := range p.curScope.vars {
		switch v := sv.(type) {
		case *vars.LVar:
			offset = types.AlignTo(offset, v.Type().Alignment())
			offset += v.Type().Size()
			v.Offset = offset
			fn.LVars = append(fn.LVars, v)
		}
	}
}

func (p *Parser) structDecl() types.Type {
	p.expect("struct")
	tag, tagExists := p.consumeID()
	if tagExists && !p.beginsWith("{") {
		if tag := p.searchStructTag(tag.Str()); tag != nil {
			return tag.ty
		}
		log.Fatalf("No such struct tag %s", tag.Str())
	}
	p.expect("{")
	var members []*types.Member
	offset, align := 0, 0
	for !p.consume("}") {
		// TODO: handle when rhs is not null
		ty, tag, _, _ := p.decl()
		offset = types.AlignTo(offset, ty.Alignment())
		members = append(members, types.NewMember(tag, offset, ty))
		offset += ty.Size()
		if align < ty.Size() {
			align = ty.Size()
		}
	}
	ty := types.NewStruct(align, members, types.AlignTo(offset, align))
	if tagExists {
		p.curScope.addStructTag(newStructTag(tag.Str(), ty))
	}
	return ty
}

func (p *Parser) enumDecl() types.Type {
	p.expect("enum")
	tag, tagExists := p.consumeID()
	if tagExists && !p.beginsWith("{") {
		if tag := p.searchEnumTag(tag.Str()); tag != nil {
			return tag.ty
		}
		log.Fatalf("No such enum tag %s", tag.Str())
	}
	t := types.NewEnum()

	p.expect("{")
	c := 0
	for {
		id := p.expectID()
		if p.consume("=") {
			c = int(p.constExpr())
		}
		p.curScope.addEnum(id.Str(), t, c)
		c++
		orig := p.Toks
		if p.consume("}") || p.consume(",") && p.consume("}") {
			break
		}
		p.Toks = orig
		p.expect(",")
	}
	if tagExists {
		p.curScope.addEnumTag(newEnumTag(tag.Str(), t))
	}
	return t
}

func (p *Parser) stmt() ast.Node {
	// handle block
	if p.consume("{") {
		var blkStmts []ast.Node
		p.spawnScope()
		for !p.consume("}") {
			blkStmts = append(blkStmts, p.stmt())
		}
		p.rewindScope()
		return ast.NewBlkNode(blkStmts)
	}

	// handle return
	if p.consume("return") {
		if p.consume(";") {
			return ast.NewRetNode(nil, p.curFnName)
		}
		node := ast.NewRetNode(p.expr(), p.curFnName)
		p.expect(";")
		return node
	}

	// handle break
	if p.consume("break") {
		p.expect(";")
		return ast.NewBreakNode()
	}

	// handle continue
	if p.consume("continue") {
		p.expect(";")
		return ast.NewContinueNode()
	}

	// handle if statement
	if p.consume("if") {
		p.expect("(")
		cond := p.expr()
		p.expect(")")
		then := p.stmt()

		var els ast.Node
		if p.consume("else") {
			els = p.stmt()
		}

		return ast.NewIfNode(cond, then, els)
	}

	// handle while statement
	if p.consume("while") {
		p.expect("(")
		cond := p.expr()
		p.expect(")")
		then := p.stmt()
		return ast.NewWhileNode(cond, then)
	}

	// handle do-while statement
	if p.consume("do") {
		then := p.stmt()
		p.expect("while")
		p.expect("(")
		cond := p.expr()
		p.expect(")")
		p.expect(";")
		return ast.NewDoWhileNode(cond, then)
	}

	// handle for statement
	if p.consume("for") {
		p.expect("(")

		var init, cond, inc, then ast.Node

		p.spawnScope()
		if !p.consume(";") {
			if p.isType() {
				t, id, rhs, _ := p.decl()
				p.curScope.addLVar(id, t)
				if rhs == nil {
					init = ast.NewNullNode()
				} else {
					a := ast.NewAssignNode(ast.NewVarNode(p.findVar(id)), rhs)
					init = ast.NewExprNode(a)
				}
			} else {
				init = ast.NewExprNode(p.expr())
				p.expect(";")
			}
		}

		if !p.consume(";") {
			cond = p.expr()
			p.expect(";")
		}

		if !p.consume(")") {
			inc = ast.NewExprNode(p.expr())
			p.expect(")")
		}

		then = p.stmt()
		p.rewindScope()
		return ast.NewForNode(init, cond, inc, then)
	}

	// handle switch statement
	if p.consume("switch") {
		p.expect("(")
		e := p.expr()
		p.expect(")")
		p.expect("{")

		var cases []*ast.CaseNode
		var dflt *ast.CaseNode
		for idx := 0; ; idx++ {
			if node, isDefault := p.switchCase(idx); node == nil {
				break
			} else {
				cases = append(cases, node)
				if isDefault {
					if dflt != nil {
						log.Fatal("Multiple definition of default clause.")
					}
					dflt = node
				}
			}
		}
		p.expect("}")
		return ast.NewSwitchNode(e, cases, dflt)
	}

	// handle variable definition
	if p.isType() {
		t, id, rhs, sc := p.decl()
		if id == "" {
			return ast.NewNullNode()
		}
		if (sc & static) != 0 {
			init := buildGVarInit(t, rhs)
			p.curScope.addGVar(true, id, t, init)
			return ast.NewNullNode()
		}
		p.curScope.addLVar(id, t)
		if rhs == nil {
			return ast.NewNullNode()
		}
		return p.storeInit(t, ast.NewVarNode(p.findVar(id)), rhs)
	}

	node := p.expr()
	p.expect(";")
	return ast.NewExprNode(node)
}

func (p *Parser) storeInit(t types.Type, dst ast.AddressableNode, rhs ast.Node) ast.Node {
	switch t := t.(type) {
	case *types.Arr:
		// TODO: clean up
		var body []ast.Node
		var ln, idx int
		_, isChar := t.Base().(*types.Char)
		isString, str := isStrNode(rhs)
		// string literal
		if isChar && isString {
			for i, r := range str {
				idx++
				addr := ast.NewDerefNode(ast.NewAddNode(dst, ast.NewNumNode(int64(i))))
				body = append(body, ast.NewExprNode(ast.NewAssignNode(addr, ast.NewNumNode(int64(r)))))
			}
			ln = len(str)
		} else {
			for i, mem := range rhs.(*ast.BlkNode).Body {
				idx++
				addr := ast.NewDerefNode(ast.NewAddNode(dst, ast.NewNumNode(int64(i))))
				body = append(body, p.storeInit(t.Base(), addr, mem))
			}
			ln = len(rhs.(*ast.BlkNode).Body)
		}

		if t.Len < 0 {
			t.Len = ln
		}

		if t, ok := t.Base().(*types.Arr); !ok {
			// zero out on initialization
			for i := idx; i < ln; i++ {
				addr := ast.NewDerefNode(ast.NewAddNode(dst, ast.NewNumNode(int64(i))))
				body = append(body, p.storeInit(t, addr, ast.NewNumNode(0)))
			}
		}

		return ast.NewBlkNode(body)
	case *types.Struct:
		var body []ast.Node
		idx := 0
		for i, mem := range rhs.(*ast.BlkNode).Body {
			idx++
			node := ast.NewMemberNode(dst, t.Members[i])
			if mem == nil {
				body = append(body, ast.NewExprNode(ast.NewAssignNode(node, ast.NewNumNode(0))))
			} else {
				body = append(body, ast.NewExprNode(ast.NewAssignNode(node, mem)))
			}
		}
		return ast.NewBlkNode(body)
	default:
		return ast.NewExprNode(ast.NewAssignNode(dst, rhs))
	}
}

func isStrNode(n ast.Node) (bool, string) {
	if v, ok := n.(*ast.VarNode); !ok {
		return false, ""
	} else if g, ok := v.Var.(*vars.GVar); !ok {
		return false, ""
	} else if t, ok := g.Type().(*types.Arr); !ok {
		return false, ""
	} else if _, ok := t.Base().(*types.Char); !ok {
		return false, ""
	} else {
		return true, g.Init.(*vars.GVarInitStr).Content
	}
}

func (p *Parser) switchCase(idx int) (node *ast.CaseNode, isDefault bool) {
	if p.consume("case") {
		n := p.constExpr()
		p.expect(":")
		var body []ast.Node
		for !p.beginsWith("case") && !p.beginsWith("default") && !p.beginsWith("}") {
			body = append(body, p.stmt())
		}
		node = ast.NewCaseNode(int(n), body, idx)
	} else if p.consume("default") {
		isDefault = true
		p.expect(":")
		var body []ast.Node
		for !p.beginsWith("case") && !p.beginsWith("default") && !p.beginsWith("}") {
			body = append(body, p.stmt())
		}
		node = ast.NewCaseNode(-1, body, idx)
	}
	return
}

func (p *Parser) expr() ast.Node {
	return p.assign()
}

func (p *Parser) assign() ast.Node {
	node := p.ternary()
	if p.consume("=") {
		node = ast.NewAssignNode(node.(ast.AddressableNode), p.assign())
	} else if p.consume("+=") {
		if _, ok := node.LoadType().(types.Pointing); ok {
			node = ast.NewBinaryNode(ast.NdPtrAddEq, node.(ast.AddressableNode), p.assign())
		} else {
			node = ast.NewBinaryNode(ast.NdAddEq, node.(ast.AddressableNode), p.assign())
		}
	} else if p.consume("-=") {
		if _, ok := node.LoadType().(types.Pointing); ok {
			node = ast.NewBinaryNode(ast.NdPtrSubEq, node.(ast.AddressableNode), p.assign())
		} else {
			node = ast.NewBinaryNode(ast.NdSubEq, node.(ast.AddressableNode), p.assign())
		}
	} else if p.consume("*=") {
		node = ast.NewBinaryNode(ast.NdMulEq, node.(ast.AddressableNode), p.assign())
	} else if p.consume("/=") {
		node = ast.NewBinaryNode(ast.NdDivEq, node.(ast.AddressableNode), p.assign())
	}
	return node
}

func (p *Parser) ternary() ast.Node {
	node := p.logOr()
	if !p.consume("?") {
		return node
	}
	l := p.expr()
	p.expect(":")
	r := p.ternary()
	return ast.NewTernaryNode(node, l, r)
}

func (p *Parser) logOr() ast.Node {
	node := p.logAnd()
	for p.consume("||") {
		node = ast.NewBinaryNode(ast.NdLogOr, node, p.logAnd())
	}
	return node
}

func (p *Parser) logAnd() ast.Node {
	node := p.bitOr()
	for p.consume("&&") {
		node = ast.NewBinaryNode(ast.NdLogAnd, node, p.bitOr())
	}
	return node
}

func (p *Parser) bitOr() ast.Node {
	node := p.bitXor()
	for p.consume("|") {
		node = ast.NewBinaryNode(ast.NdBitOr, node, p.bitXor())
	}
	return node
}

func (p *Parser) bitXor() ast.Node {
	node := p.bitAnd()
	for p.consume("^") {
		node = ast.NewBinaryNode(ast.NdBitXor, node, p.bitXor())
	}
	return node
}

func (p *Parser) bitAnd() ast.Node {
	node := p.equality()
	for p.consume("&") {
		node = ast.NewBinaryNode(ast.NdBitAnd, node, p.equality())
	}
	return node
}

func (p *Parser) equality() ast.Node {
	node := p.relational()
	for {
		if p.consume("==") {
			node = ast.NewBinaryNode(ast.NdEq, node, p.relational())
		} else if p.consume("!=") {
			node = ast.NewBinaryNode(ast.NdNeq, node, p.relational())
		} else {
			return node
		}
	}
}

func (p *Parser) relational() ast.Node {
	node := p.shift()
	for {
		if p.consume("<=") {
			node = ast.NewBinaryNode(ast.NdLeq, node, p.shift())
		} else if p.consume(">=") {
			node = ast.NewBinaryNode(ast.NdGeq, node, p.shift())
		} else if p.consume("<") {
			node = ast.NewBinaryNode(ast.NdLt, node, p.shift())
		} else if p.consume(">") {
			node = ast.NewBinaryNode(ast.NdGt, node, p.shift())
		} else {
			return node
		}
	}
}

func (p *Parser) shift() ast.Node {
	node := p.addSub()
	for {
		if p.consume("<<") {
			node = ast.NewBinaryNode(ast.NdShl, node, p.shift())
		} else if p.consume(">>") {
			node = ast.NewBinaryNode(ast.NdShr, node, p.shift())
		} else if p.consume("<<=") {
			node = ast.NewBinaryNode(ast.NdShlEq, node, p.shift())
		} else if p.consume(">>=") {
			node = ast.NewBinaryNode(ast.NdShrEq, node, p.shift())
		} else {
			return node
		}
	}
}

func (p *Parser) addSub() ast.Node {
	node := p.mulDiv()
	for {
		if p.consume("+") {
			node = ast.NewAddNode(node, p.mulDiv())
		} else if p.consume("-") {
			node = ast.NewSubNode(node, p.mulDiv())
		} else {
			return node
		}
	}
}

func (p *Parser) mulDiv() ast.Node {
	node := p.cast()
	for {
		if p.consume("*") {
			node = ast.NewBinaryNode(ast.NdMul, node, p.cast())
		} else if p.consume("/") {
			node = ast.NewBinaryNode(ast.NdDiv, node, p.cast())
		} else {
			return node
		}
	}
}

func (p *Parser) cast() ast.Node {
	orig := p.Toks
	if p.consume("(") {
		if p.isType() {
			t, _, _ := p.baseType()
			for p.consume("*") {
				t = types.NewPtr(t)
			}
			p.expect(")")
			return ast.NewCastNode(p.cast(), t)
		}
		p.Toks = orig
	}
	return p.unary()
}

func (p *Parser) unary() ast.Node {
	if p.consume("+") {
		return p.cast()
	}
	if p.consume("-") {
		return ast.NewSubNode(ast.NewNumNode(0), p.cast())
	}
	if p.consume("*") {
		return ast.NewDerefNode(p.cast())
	}
	if p.consume("&") {
		return ast.NewAddrNode(p.cast().(ast.AddressableNode))
	}
	if p.consume("!") {
		return ast.NewNotNode(p.cast())
	}
	if p.consume("~") {
		return ast.NewBitNotNode(p.cast())
	}
	if p.consume("++") {
		return ast.NewIncNode(p.unary().(ast.AddressableNode), true)
	}
	if p.consume("--") {
		return ast.NewDecNode(p.unary().(ast.AddressableNode), true)
	}
	return p.postfix()
}

func (p *Parser) postfix() ast.Node {
	node := p.primary()
	for {
		if p.consume("[") {
			add := ast.NewAddNode(node, p.expr())
			node = ast.NewDerefNode(add)
			p.expect("]")
			continue
		}
		if p.consume(".") {
			if s, ok := node.LoadType().(*types.Struct); ok {
				mem := s.FindMember(p.expectID().Str())
				node = ast.NewMemberNode(node.(ast.AddressableNode), mem)
				continue
			}
			log.Fatalf("Expected struct but got %T", node.LoadType())
		}
		if p.consume("->") {
			if t, ok := node.LoadType().(*types.Ptr); ok {
				mem := t.Base().(*types.Struct).FindMember(p.expectID().Str())
				node = ast.NewMemberNode(ast.NewDerefNode(node.(ast.AddressableNode)), mem)
				continue
			}
			log.Fatalf("Expected pointer but got %T", node.LoadType())
		}
		if p.consume("++") {
			node = ast.NewIncNode(node.(ast.AddressableNode), false)
			continue
		}
		if p.consume("--") {
			node = ast.NewDecNode(node.(ast.AddressableNode), false)
			continue
		}
		return node
	}
}

func (p *Parser) stmtExpr() ast.Node {
	// "(" and "{" is already read.
	p.spawnScope()
	body := make([]ast.Node, 0)
	body = append(body, p.stmt())
	for !p.consume("}") {
		body = append(body, p.stmt())
	}
	p.expect(")")
	if ex, ok := body[len(body)-1].(*ast.ExprNode); !ok {
		log.Fatal("Statement expression returning void is not supported")
	} else {
		body[len(body)-1] = ex.Body
	}
	p.rewindScope()
	return ast.NewStmtExprNode(body)
}

func (p *Parser) primary() ast.Node {
	if p.consume("(") {
		if p.consume("{") {
			return p.stmtExpr()
		}
		node := p.expr()
		p.expect(")")
		return node
	}

	if p.consume("sizeof") {
		orig := p.Toks
		if p.consume("(") {
			if p.isType() {
				base, _, _ := p.baseType()
				n := base.Size()
				p.expect(")")
				return ast.NewNumNode(int64(n))
			}
			p.Toks = orig
		}
		return ast.NewNumNode(int64(p.unary().LoadType().Size()))
	}

	if id, isID := p.consumeID(); isID {
		id := id.Str()
		if p.consume("(") {
			var t types.Type
			if fn, ok := p.searchVar(id).(*vars.GVar); ok {
				t = fn.Type().(*types.Fn).RetTy
			} else {
				t = types.NewInt()
			}
			var params []ast.Node
			if p.consume(")") {
				return ast.NewFnCallNode(id, params, t)
			}
			params = append(params, p.expr())
			for p.consume(",") {
				params = append(params, p.expr())
			}
			p.expect(")")
			return ast.NewFnCallNode(id, params, t)
		}

		switch v := p.findVar(id).(type) {
		case *vars.Enum:
			return ast.NewNumNode(int64(v.Val))
		case *vars.LVar, *vars.GVar:
			return ast.NewVarNode(v)
		default:
			log.Fatalf("Unhandled case of vars.Var in primary: %T", p.findVar(id))
		}
	}

	if strTok, isStr := p.consumeStr(); isStr {
		init := vars.NewGVarInitStr(strTok.Str())
		s := vars.NewGVar(true, newGVarLabel(), types.NewArr(types.NewChar(), strTok.Len()), init)
		p.Ast.GVars = append(p.Ast.GVars, s)
		return ast.NewVarNode(s)
	}

	return ast.NewNumNode(int64(p.expectNum().Val))
}
