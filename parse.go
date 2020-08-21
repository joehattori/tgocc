package main

import (
	"fmt"
	"log"
	"strings"
	"unicode/utf8"
)

func (t *tokenized) popToks() {
	t.toks = t.toks[1:]
}

func (t *tokenized) consume(str string) bool {
	if r, isReserved := t.toks[0].(*reservedTok); isReserved &&
		r.len == len(str) &&
		strings.HasPrefix(r.str, str) {
		t.popToks()
		return true
	}
	return false
}

func (t *tokenized) consumeID() (string, bool) {
	cur := t.toks[0]
	id := idRegexp.FindString(cur.getStr())
	if _, isID := cur.(*idTok); isID && id != "" {
		t.popToks()
		return id, true
	}
	return "", false
}

func (t *tokenized) consumeStr() (string, bool) {
	if s, isStr := t.toks[0].(*strTok); isStr {
		t.popToks()
		return s.content, true
	}
	return "", false
}

func (t *tokenized) expect(str string) {
	cur := t.toks[0]
	if r, isReserved := cur.(*reservedTok); isReserved &&
		r.len == len(str) && strings.HasPrefix(cur.getStr(), str) {
		t.popToks()
		return
	}
	log.Fatalf("%s was expected but got %s", str, cur.getStr())
}

func (t *tokenized) expectID() string {
	cur := t.toks[0]
	id := idRegexp.FindString(cur.getStr())
	if _, isID := cur.(*idTok); isID && id != "" {
		t.popToks()
		return id
	}
	log.Fatalf("ID was expected but got %s", cur.getStr())
	return ""
}

func (t *tokenized) expectNum() int {
	cur := t.toks[0]
	if n, isNum := cur.(*numTok); isNum {
		t.popToks()
		return n.val
	}
	log.Fatalf("Number was expected but got %s", cur.getStr())
	return -1
}

var tyMap = map[string]ty{
	"int":  newTyInt(),
	"char": newTyChar(),
}

func (t *tokenized) expectType() ty {
	cur := t.toks[0]
	if r, isReserved := cur.(*reservedTok); isReserved {
		for key, val := range tyMap {
			if strings.HasPrefix(r.str, key) {
				t.popToks()
				return val
			}
		}
		log.Fatalf("Unexpected type %s", r.str)
	}
	log.Fatalf("tkReserved was expected but got %d %s", cur, cur.getStr())
	return nil
}

func (t *tokenized) peekType() bool {
	if r, isReserved := t.toks[0].(*reservedTok); isReserved {
		for key := range tyMap {
			if strings.HasPrefix(r.str, key) {
				return true
			}
		}
	}
	return false
}

var gVarLabelCount int

func newGVarLabel() string {
	ret := fmt.Sprintf(".L.data.%d", gVarLabelCount)
	gVarLabelCount++
	return ret
}

// program    = (function | globalVar)*
// function   = ty ident ("(" (ident ("," ident)* )? ")") "{" stmt* "}"
// globalVar  = varDecl
// stmt       = expr ";"
//				| "{" stmt* "}"
//			    | "return" expr ";"
//				| "if" "(" expr ")" stmt ("else" stmt) ?
//				| "while" "(" expr ")" stmt
//				| "for" "(" expr? ";" expr? ";" expr? ")" stmt
//				| varDecl
// varDecl    = ty ident ("[" expr "]")* "=" expr ;"
// expr       = assign
// assign     = equality ("=" assign) ?
// equality   = relational ("==" relational | "!=" relational)*
// relational = add ("<" add | "<=" add | ">" add | ">=" add)*
// add        = mul ("+" mul | "-" mul)*
// mul        = unary ("*" unary | "/" unary)*
// unary      = "sizeof" unary | ("+" | "-" | "*" | "&")? unary | postfix
// postfix    = primary ("[" expr "]")*
// primary    = num | str | ident ("(" (expr ("," expr)* )? ")")? | "(" expr ")"

func (t *tokenized) isFunction() bool {
	orig := t.toks
	ty := t.expectType()
	for t.consume("*") {
		ty = &tyPtr{to: ty}
	}
	_, isID := t.consumeID()
	ret := isID && t.consume("(")
	t.toks = orig
	return ret
}

func (t *tokenized) readVarSuffix(ty ty) ty {
	if t.consume("[") {
		l := t.expectNum()
		t.expect("]")
		ty = t.readVarSuffix(ty)
		return newTyArr(ty, l)
	}
	return ty
}

func (t *tokenized) varDecl() (id string, ty ty, rhs node) {
	ty = t.expectType()
	for t.consume("*") {
		ty = &tyPtr{to: ty}
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

func (t *tokenized) parse() {
	ast := t.res
	for {
		if _, isEOF := t.toks[0].(*eofTok); isEOF {
			break
		}
		if t.isFunction() {
			ast.fns = append(ast.fns, t.function())
		} else {
			id, ty, _ := t.varDecl()
			g := newGVar(id, ty, nil)
			ast.gvars = append(ast.gvars, g)
		}
	}
}

func (t *tokenized) function() *fnNode {
	ty := t.expectType()
	for t.consume("*") {
		ty = newTyPtr(ty)
	}
	funcName := t.expectID()
	t.expect("(")
	fn := &fnNode{name: funcName, ty: ty}
	t.curFn = fn

	isFirstArg := true
	for !t.consume(")") {
		if !isFirstArg {
			t.expect(",")
		}
		isFirstArg = false

		ty := t.expectType()
		for t.consume("*") {
			ty = &tyPtr{to: ty}
		}
		s := t.expectID()
		t.buildLVarNode(s, ty, true)
	}
	t.expect("{")
	for !t.consume("}") {
		fn.body = append(fn.body, t.stmt())
	}
	return fn
}

func (t *tokenized) stmt() node {
	// handle block
	if t.consume("{") {
		var blkStmts []node
		for {
			if t.consume("}") {
				return newBlkNode(blkStmts)
			}
			blkStmts = append(blkStmts, t.stmt())
		}
	}

	// handle return
	if t.consume("return") {
		node := newRetNode(t.expr(), t.curFn.name)
		t.expect(";")
		return node
	}

	// handle if statement
	if t.consume("if") {
		t.expect("(")
		cond := t.expr()
		t.expect(")")
		then := t.stmt()

		var els node
		if t.consume("else") {
			els = t.stmt()
		}

		return newIfNode(cond, then, els)
	}

	// handle while statement
	if t.consume("while") {
		t.expect("(")
		condNode := t.expr()
		t.expect(")")
		thenNode := t.stmt()
		return newWhileNode(condNode, thenNode)
	}

	// handle for statement
	if t.consume("for") {
		t.expect("(")

		var init, cond, inc, then node

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
		return newForNode(init, cond, inc, then)
	}

	// handle variable definition
	if t.peekType() {
		id, ty, rhs := t.varDecl()
		lv := t.buildLVarNode(id, ty, false)
		if rhs == nil {
			return lv
		}
		return newAssignNode(newVarNode(t.findVar(id)).(addressableNode), rhs)
	}

	node := t.expr()
	t.expect(";")
	return newExprNode(node)
}

func (t *tokenized) expr() node {
	return t.assign()
}

func (t *tokenized) assign() node {
	node := t.equality()

	if t.consume("=") {
		assignNode := t.assign()
		node = newAssignNode(node.(addressableNode), assignNode)
	}
	return node
}

func (t *tokenized) equality() node {
	node := t.relational()
	for {
		if t.consume("==") {
			node = newArithNode(ndEq, node, t.relational())
		} else if t.consume("!=") {
			node = newArithNode(ndNeq, node, t.relational())
		} else {
			return node
		}
	}
}

func (t *tokenized) relational() node {
	node := t.addSub()
	for {
		if t.consume("<=") {
			node = newArithNode(ndLeq, node, t.addSub())
		} else if t.consume(">=") {
			node = newArithNode(ndGeq, node, t.addSub())
		} else if t.consume("<") {
			node = newArithNode(ndLt, node, t.addSub())
		} else if t.consume(">") {
			node = newArithNode(ndGt, node, t.addSub())
		} else {
			return node
		}
	}
}

func (t *tokenized) addSub() node {
	node := t.mulDiv()
	for {
		if t.consume("+") {
			node = newAddNode(node, t.mulDiv())
		} else if t.consume("-") {
			node = newSubNode(node, t.mulDiv())
		} else {
			return node
		}
	}
}

func (t *tokenized) mulDiv() node {
	node := t.unary()
	for {
		if t.consume("*") {
			node = newArithNode(ndMul, node, t.unary())
		} else if t.consume("/") {
			node = newArithNode(ndDiv, node, t.unary())
		} else {
			return node
		}
	}
}

func (t *tokenized) unary() node {
	if t.consume("sizeof") {
		return newNumNode(t.unary().loadType().size())
	}
	if t.consume("+") {
		return t.unary()
	}
	if t.consume("-") {
		node := t.unary()
		return newSubNode(newNumNode(0), node)
	}
	if t.consume("*") {
		node := t.unary()
		return newDerefNode(node)
	}
	if t.consume("&") {
		node := t.unary()
		return newAddrNode(node.(addressableNode))
	}
	return t.postfix()
}

func (t *tokenized) postfix() node {
	node := t.primary()
	for t.consume("[") {
		add := newAddNode(node, t.expr())
		node = newDerefNode(add)
		t.expect("]")
	}
	return node
}

func (t *tokenized) primary() node {
	if t.consume("(") {
		node := t.expr()
		t.expect(")")
		return node
	}

	if id, isID := t.consumeID(); isID {
		if t.consume("(") {
			var params []node
			if t.consume(")") {
				return newFnCallNode(id, params)
			}
			params = append(params, t.expr())
			for t.consume(",") {
				params = append(params, t.expr())
			}
			t.expect(")")
			return newFnCallNode(id, params)
		}
		return newVarNode(t.findVar(id))
	}

	if str, isStr := t.consumeStr(); isStr {
		s := newGVar(newGVarLabel(), newTyArr(newTyChar(), utf8.RuneCountInString(str)), str)
		t.res.gvars = append(t.res.gvars, s)
		return newVarNode(s)
	}

	return newNumNode(t.expectNum())
}
