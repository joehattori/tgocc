package main

import (
	"fmt"
	"regexp"
	"strings"
)

func (t *Tokenized) popToks() {
	t.toks = t.toks[1:]
}

func (t *Tokenized) consume(str string) bool {
	cur := t.toks[0]
	if cur.kind != tkReserved || cur.length != len(str) || !strings.HasPrefix(cur.str, str) {
		return false
	}
	t.popToks()
	return true
}

func (t *Tokenized) consumeID() (string, bool) {
	cur := t.toks[0]
	reg := regexp.MustCompile(idRegexp)
	varName := reg.FindString(cur.str)
	if cur.kind != tkID || varName == "" {
		return "", false
	}
	t.popToks()
	return varName, true
}

func (t *Tokenized) consumeSizeOf() bool {
	cur := t.toks[0]
	if cur.kind == tkSizeOf && strings.HasPrefix(cur.str, "sizeof") {
		t.popToks()
		return true
	}
	return false
}

func (t *Tokenized) expect(str string) {
	cur := t.toks[0]
	if cur.kind != tkReserved || cur.length != len(str) || !strings.HasPrefix(cur.str, str) {
		panic(fmt.Sprintf("%s was expected but got %s", str, cur.str))
	}
	t.popToks()
}

func (t *Tokenized) expectID() string {
	cur := t.toks[0]
	reg := regexp.MustCompile(idRegexp)
	varName := reg.FindString(cur.str)
	if cur.kind != tkID || varName == "" {
		panic(fmt.Sprintf("ID was expected but got %s", cur.str))
	}
	t.popToks()
	return varName
}

func (t *Tokenized) expectNum() int {
	cur := t.toks[0]
	if cur.kind != tkNum {
		panic(fmt.Sprintf("Number was expected but got %s", cur.str))
	}
	t.popToks()
	return cur.val
}

var tyArr = [...]struct {
	string
	Type
}{
	{"int", &TyInt{}},
	{"char", &TyChar{}},
}

func (t *Tokenized) expectType() Type {
	cur := t.toks[0]
	if cur.kind != tkReserved {
		panic(fmt.Sprintf("tkReserved was expected but got %d %s", cur.kind, cur.str))
	}
	for _, a := range tyArr {
		if strings.HasPrefix(cur.str, a.string) {
			t.popToks()
			return a.Type
		}
	}
	panic(fmt.Sprintf("Unexpected type %s", cur.str))
}

func (t *Tokenized) peekType() bool {
	cur := t.toks[0]
	if cur.kind != tkReserved {
		return false
	}
	for _, a := range tyArr {
		if strings.HasPrefix(cur.str, a.string) {
			return true
		}
	}
	return false
}

// program    = (function | globalVar)*
// function   = Type ident ("(" (ident ("," ident)* )? ")") "{" stmt* "}"
// globalVar  = varDecl
// stmt       = expr ";"
//				| "{" stmt* "}"
//			    | "return" expr ";"
//				| "if" "(" expr ")" stmt ("else" stmt) ?
//				| "while" "(" expr ")" stmt
//				| "for" "(" expr? ";" expr? ";" expr? ")" stmt
//				| varDecl
// varDecl    = Type ident ("[" expr "]")* ";"
// expr       = assign
// assign     = equality ("=" assign) ?
// equality   = relational ("==" relational | "!=" relational)*
// relational = add ("<" add | "<=" add | ">" add | ">=" add)*
// add        = mul ("+" mul | "-" mul)*
// mul        = unary ("*" unary | "/" unary)*
// unary      = "sizeof" unary | ("+" | "-" | "*" | "&")? unary | postfix
// postfix    = primary ("[" expr "]")*
// primary    = num | ident ("(" (expr ("," expr)* )? ")")? | "(" expr ")"

func (t *Tokenized) isFunction() bool {
	orig := t.toks
	ty := t.expectType()
	for t.consume("*") {
		ty = &TyPtr{to: ty}
	}
	_, isID := t.consumeID()
	ret := isID && t.consume("(")
	t.toks = orig
	return ret
}

func (t *Tokenized) readVarSuffix(ty Type) Type {
	if t.consume("[") {
		l := t.expectNum()
		t.expect("]")
		ty = t.readVarSuffix(ty)
		return &TyArr{of: ty, len: l}
	}
	return ty
}

func (t *Tokenized) varDecl() (id string, ty Type, rhs Node) {
	ty = t.expectType()
	for t.consume("*") {
		ty = &TyPtr{to: ty}
	}
	id = t.expectID()
	ty = t.readVarSuffix(ty)
	if t.consume(";") {
		return
	}
	t.expect("=")
	rhs = t.expr()
	t.expect(";")
	return
}

func (t *Tokenized) parse() {
	ast := t.res
	for t.toks[0].kind != tkEOF {
		if t.isFunction() {
			ast.fns = append(ast.fns, t.function())
		} else {
			id, ty, _ := t.varDecl()
			g := NewGVar(id, ty)
			ast.gvars = append(ast.gvars, g)
		}
	}
}

func (t *Tokenized) function() *FnNode {
	ty := t.expectType()
	for t.consume("*") {
		ty = &TyPtr{to: ty}
	}
	funcName := t.expectID()
	t.expect("(")
	fn := &FnNode{name: funcName, ty: ty}
	t.curFn = fn

	isFirstArg := true
	for !t.consume(")") {
		if !isFirstArg {
			t.expect(",")
		}
		isFirstArg = false

		ty := t.expectType()
		for t.consume("*") {
			ty = &TyPtr{to: ty}
		}
		s := t.expectID()
		t.BuildLVarNode(s, ty, true)
	}
	t.expect("{")
	for !t.consume("}") {
		fn.body = append(fn.body, t.stmt())
	}
	return fn
}

func (t *Tokenized) stmt() Node {
	// handle block
	if t.consume("{") {
		var blkStmts []Node
		for {
			if t.consume("}") {
				return NewBlkNode(blkStmts)
			}
			blkStmts = append(blkStmts, t.stmt())
		}
	}

	// handle return
	if t.consume("return") {
		node := NewRetNode(t.expr(), t.curFn.name)
		t.expect(";")
		return node
	}

	// handle if statement
	if t.consume("if") {
		t.expect("(")
		cond := t.expr()
		t.expect(")")
		then := t.stmt()

		var els Node
		if t.consume("else") {
			els = t.stmt()
		}

		return NewIfNode(cond, then, els)
	}

	// handle while statement
	if t.consume("while") {
		t.expect("(")
		condNode := t.expr()
		t.expect(")")
		thenNode := t.stmt()
		return NewWhileNode(condNode, thenNode)
	}

	// handle for statement
	if t.consume("for") {
		t.expect("(")

		var init, cond, inc, then Node

		if !t.consume(";") {
			init = t.expr()
			t.expect(";")
		}

		if !t.consume(";") {
			cond = t.expr()
			t.expect(";")
		}

		if !t.consume(")") {
			inc = t.expr()
			t.expect(")")
		}

		then = t.stmt()
		return NewForNode(init, cond, inc, then)
	}

	// handle variable definition
	if t.peekType() {
		id, ty, rhs := t.varDecl()
		lv := t.BuildLVarNode(id, ty, false)
		if rhs == nil {
			return lv
		}
		return NewAssignNode(NewVarNode(t.FindVar(id)).(AddressableNode), rhs)
	}

	node := t.expr()
	t.expect(";")
	return NewExprNode(node)
}

func (t *Tokenized) expr() Node {
	return t.assign()
}

func (t *Tokenized) assign() Node {
	node := t.equality()

	if t.consume("=") {
		assignNode := t.assign()
		node = NewAssignNode(node.(AddressableNode), assignNode)
	}
	return node
}

func (t *Tokenized) equality() Node {
	node := t.relational()
	for {
		if t.consume("==") {
			node = NewArithNode(ndEq, node, t.relational())
		} else if t.consume("!=") {
			node = NewArithNode(ndNeq, node, t.relational())
		} else {
			return node
		}
	}
}

func (t *Tokenized) relational() Node {
	node := t.addSub()
	for {
		if t.consume("<=") {
			node = NewArithNode(ndLeq, node, t.addSub())
		} else if t.consume(">=") {
			node = NewArithNode(ndGeq, node, t.addSub())
		} else if t.consume("<") {
			node = NewArithNode(ndLt, node, t.addSub())
		} else if t.consume(">") {
			node = NewArithNode(ndGt, node, t.addSub())
		} else {
			return node
		}
	}
}

func (t *Tokenized) addSub() Node {
	node := t.mulDiv()
	for {
		if t.consume("+") {
			node = NewAddNode(node, t.mulDiv())
		} else if t.consume("-") {
			node = NewSubNode(node, t.mulDiv())
		} else {
			return node
		}
	}
}

func (t *Tokenized) mulDiv() Node {
	node := t.unary()
	for {
		if t.consume("*") {
			node = NewArithNode(ndMul, node, t.unary())
		} else if t.consume("/") {
			node = NewArithNode(ndDiv, node, t.unary())
		} else {
			return node
		}
	}
}

func (t *Tokenized) unary() Node {
	if t.consumeSizeOf() {
		return NewNumNode(t.unary().loadType().size())
	}
	if t.consume("+") {
		return t.unary()
	}
	if t.consume("-") {
		node := t.unary()
		return NewSubNode(NewNumNode(0), node)
	}
	if t.consume("*") {
		node := t.unary()
		return NewDerefNode(node)
	}
	if t.consume("&") {
		node := t.unary()
		return NewAddrNode(node.(AddressableNode))
	}
	return t.postfix()
}

func (t *Tokenized) postfix() Node {
	node := t.primary()
	for t.consume("[") {
		add := NewAddNode(node, t.expr())
		node = NewDerefNode(add)
		t.expect("]")
	}
	return node
}

func (t *Tokenized) primary() Node {
	if t.consume("(") {
		node := t.expr()
		t.expect(")")
		return node
	}

	if id, isID := t.consumeID(); isID {
		if t.consume("(") {
			var params []Node
			if t.consume(")") {
				return NewFnCallNode(id, params)
			}
			params = append(params, t.expr())
			for t.consume(",") {
				params = append(params, t.expr())
			}
			t.expect(")")
			return NewFnCallNode(id, params)
		}
		return NewVarNode(t.FindVar(id))
	}

	return NewNumNode(t.expectNum())
}
