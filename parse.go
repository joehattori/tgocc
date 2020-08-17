package main

import (
	"fmt"
	"regexp"
	"strings"
)

// LVar represents local variable. `offset` is the offset from rbp
type LVar struct {
	name   string
	offset int
	ty     Type
}

// Ast is an array of nodes which represents whole program input (technically not a tree, but let's call this Ast still.)
type Ast struct {
	nodes []Node
}

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
	reg := regexp.MustCompile(`^[a-zA-Z]+[\w_]*`)
	varName := reg.FindString(cur.str)
	if cur.kind != tkID || varName == "" {
		return "", false
	}
	t.popToks()
	return varName, true
}

func (t *Tokenized) consumeType() (Type, bool) {
	arr := []struct {
		string
		Type
	}{{"int", &TyInt{}}}
	for _, a := range arr {
		if t.consume(a.string) {
			return a.Type, true
		}
	}
	return nil, false
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
	reg := regexp.MustCompile(`^[a-zA-Z]+[\w_]*`)
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

func (t *Tokenized) expectType() Type {
	arr := []struct {
		string
		Type
	}{{"int", &TyInt{}}}
	cur := t.toks[0]
	if cur.kind != tkReserved {
		panic(fmt.Sprintf("tkReserved was expected but got %d %s", cur.kind, cur.str))
	}
	for _, a := range arr {
		if strings.HasPrefix(cur.str, a.string) {
			t.popToks()
			return a.Type
		}
	}
	panic(fmt.Sprintf("Unexpected type %s", cur.str))
}

// program    = function*
// function   = ident ("(" (ident ("," ident)* )? ")") "{" stmt* "}"
// stmt       = expr ";"
//				| "{" stmt* "}"
//			    | "return" expr ";"
//				| "if" "(" expr ")" stmt ("else" stmt) ?
//				| "while" "(" expr ")" stmt
//				| "for" "(" expr? ";" expr? ";" expr? ")" stmt
//				| "int" ident ";"
// expr       = assign
// assign     = equality ("=" assign) ?
// equality   = relational ("==" relational | "!=" relational)*
// relational = add ("<" add | "<=" add | ">" add | ">=" add)*
// add        = mul ("+" mul | "-" mul)*
// mul        = unary ("*" unary | "/" unary)*
// unary      = ("+" | "-" | "*" | "&")? primary
// primary    = num | ident ("(" (expr ("," expr)* )? ")")? | "(" expr ")"

type opKind struct {
	str  string
	kind nodeKind
}

func (t *Tokenized) parse() *Ast {
	var ast Ast
	for t.toks[0].kind != tkEOF {
		ast.nodes = append(ast.nodes, t.function())
	}
	return &ast
}

func (t *Tokenized) function() Node {
	ty := t.expectType()
	for {
		if !t.consume("*") {
			break
		}
		ty = &TyPtr{to: ty}
	}
	funcName := t.expectID()
	t.expect("(")

	fn := &FnNode{name: funcName, ty: ty}
	t.curFn = fn

	isFirstArg := true
	for {
		if t.consume(")") {
			break
		}

		if !isFirstArg {
			t.expect(",")
		}
		isFirstArg = false

		ty := t.expectType()
		s := t.expectID()
		arg := &LVar{name: s, offset: 8 * (len(fn.args) + 1), ty: ty}
		fn.args = append(fn.args, arg)
		fn.lvars = append(fn.lvars, arg)
	}
	t.expect("{")
	for {
		if t.consume("}") {
			break
		}
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
	if ty, isTy := t.consumeType(); isTy {
		for {
			if !t.consume("*") {
				break
			}
			ty = &TyPtr{to: ty}
		}
		id := t.expectID()
		t.expect(";")
		return t.curFn.NewLVarNode(id, ty)
	}

	node := t.expr()
	t.expect(";")
	return node
}

func (t *Tokenized) expr() Node {
	return t.assign()
}

func (t *Tokenized) assign() Node {
	node := t.equality()

	if t.consume("=") {
		assignNode := t.assign()
		node = NewAssignNode(node, assignNode)
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
			node = NewArithNode(ndAdd, node, t.mulDiv())
		} else if t.consume("-") {
			node = NewArithNode(ndSub, node, t.mulDiv())
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
	if t.consume("+") {
		return t.primary()
	}
	if t.consume("-") {
		node := t.primary()
		return NewArithNode(ndSub, NewNumNode(0), node)
	}
	if t.consume("*") {
		node := t.unary()
		return NewDerefNode(node)
	}
	if t.consume("&") {
		node := t.unary()
		return NewAddrNode(node.(*LVarNode))
	}
	return t.primary()
}

func (t *Tokenized) primary() Node {
	if t.consume("(") {
		node := t.expr()
		t.expect(")")
		return node
	}

	if id, isID := t.consumeID(); isID {
		if t.consume("(") {
			var args []Node
			if t.consume(")") {
				return NewFuncCallNode(id, args)
			}
			args = append(args, t.expr())
			for {
				if !t.consume(",") {
					break
				}
				args = append(args, t.expr())
			}
			t.expect(")")
			return NewFuncCallNode(id, args)
		}
		return t.curFn.FindLVarNode(id)
	}

	return NewNumNode(t.expectNum())
}
