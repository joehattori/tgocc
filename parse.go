package main

import (
	"fmt"
	"regexp"
	"strings"
)

//LVars represents the array of local variables
var LVars []*LVar

// LVar represents local variable. `offset` is the offset from rbp
type LVar struct {
	name   string
	offset int
}

// Code is an array of nodes which represents whole program input
var Code []Node

// TODO: move toks to instance method
// TODO: define struct program that contains token and lvar
func consume(toks []*Token, str string) ([]*Token, bool) {
	cur := toks[0]
	if cur.kind != tkReserved || cur.length != len(str) || !strings.HasPrefix(cur.str, str) {
		return toks, false
	}
	return toks[1:], true
}

func consumeID(toks []*Token) ([]*Token, string, bool) {
	cur := toks[0]
	reg := regexp.MustCompile(`^[a-zA-Z]+[\w_]*`)
	varName := reg.FindString(cur.str)
	if cur.kind != tkID || varName == "" {
		return toks, "", false
	}
	return toks[1:], varName, true
}

func expect(toks []*Token, str string) []*Token {
	cur := toks[0]
	if cur.kind != tkReserved || cur.length != len(str) || !strings.HasPrefix(cur.str, str) {
		panic(fmt.Sprintf("%s was expected but got %s", str, cur.str))
	}
	return toks[1:]
}

func expectID(toks []*Token) ([]*Token, string) {
	cur := toks[0]
	reg := regexp.MustCompile(`^[a-zA-Z]+[\w_]*`)
	varName := reg.FindString(cur.str)
	if cur.kind != tkID || varName == "" {
		panic(fmt.Sprintf("ID was expected but got %s", cur.str))
	}
	return toks[1:], varName
}

func expectNum(toks []*Token) ([]*Token, int) {
	cur := toks[0]
	if cur.kind != tkNum {
		panic("Number was expected")
	}
	return toks[1:], cur.val
}

// program    = function*
// function   = ident ("(" (ident ("," ident)* )? ")") "{" stmt* "}"
// stmt       = expr ";"
//				| "{" stmt* "}"
//			    | "return" expr ";"
//				| "if" "(" expr ")" stmt ("else" stmt) ?
//				| "while" "(" expr ")" stmt
//				| "for" "(" expr? ";" expr? ";" expr? ")" stmt
// expr       = assign
// assign     = equality ("=" assign) ?
// equality   = relational ("==" relational | "!=" relational)*
// relational = add ("<" add | "<=" add | ">" add | ">=" add)*
// add        = mul ("+" mul | "-" mul)*
// mul        = unary ("*" unary | "/" unary)*
// unary      = ("+" | "-")? primary
// primary    = num | ident ("(" (expr ("," expr)* )? ")")? | "(" expr ")"

type opKind struct {
	str  string
	kind nodeKind
}

// Program parses tokens and fills up Code
func Program(toks []*Token) {
	var node Node
	for toks[0].kind != tkEOF {
		toks, node = function(toks)
		Code = append(Code, node)
	}
}

func function(toks []*Token) ([]*Token, Node) {
	LVars = nil
	toks, funcName := expectID(toks)
	toks = expect(toks, "(")
	var body []Node
	var args []*LVar
	isFirstArg := true
	for {
		var isRPar bool
		if toks, isRPar = consume(toks, ")"); isRPar {
			break
		}

		if !isFirstArg {
			toks = expect(toks, ",")
		}
		isFirstArg = false
		var s string
		toks, s = expectID(toks)
		arg := &LVar{name: s, offset: 8 * (len(args) + 1)}
		args = append(args, arg)
		LVars = append(LVars, arg)
	}
	toks = expect(toks, "{")
	var node Node
	var isLBracket bool
	for {
		if toks, isLBracket = consume(toks, "}"); isLBracket {
			break
		}
		toks, node = stmt(toks)
		body = append(body, node)
	}
	return toks, NewFuncDefNode(funcName, args, body, 8*len(LVars))
}

func stmt(toks []*Token) ([]*Token, Node) {
	var node Node

	// handle block
	if newToks, isLBracket := consume(toks, "{"); isLBracket {
		var blkStmts []Node
		var isRBracket bool
		toks = newToks
		for {
			toks, isRBracket = consume(toks, "}")
			if isRBracket {
				return toks, NewBlkNode(blkStmts)
			}
			toks, node = stmt(toks)
			blkStmts = append(blkStmts, node)
		}
	}

	// handle return
	if newToks, isReturn := consume(toks, "return"); isReturn {
		toks, node = expr(newToks)
		node = NewRetNode(node)
		toks = expect(toks, ";")
		return toks, node
	}

	// handle if statement
	if newToks, isIf := consume(toks, "if"); isIf {
		var cond, then, els Node
		toks = expect(newToks, "(")
		toks, cond = expr(toks)
		toks = expect(toks, ")")
		toks, then = stmt(toks)

		if newToks, isElse := consume(toks, "else"); isElse {
			toks, node = stmt(newToks)
			els = node
		}

		return toks, NewIfNode(cond, then, els)
	}

	// handle while statement
	if newToks, isWhile := consume(toks, "while"); isWhile {
		var condNode, thenNode Node
		toks = expect(newToks, "(")
		toks, condNode = expr(toks)
		toks = expect(toks, ")")
		toks, thenNode = stmt(toks)
		return toks, NewWhileNode(condNode, thenNode)
	}

	// handle for statement
	if newToks, isFor := consume(toks, "for"); isFor {
		var init, cond, inc, then Node
		toks = expect(newToks, "(")

		toks, isSc := consume(toks, ";")
		if !isSc {
			toks, init = expr(toks)
			toks = expect(toks, ";")
		}

		toks, isSc = consume(toks, ";")
		if !isSc {
			toks, cond = expr(toks)
			toks = expect(toks, ";")
		}

		toks, isRPar := consume(toks, ")")
		if !isRPar {
			toks, inc = expr(toks)
			toks = expect(toks, ")")
		}

		toks, then = stmt(toks)
		return toks, NewForNode(init, cond, inc, then)
	}

	toks, node = expr(toks)
	toks = expect(toks, ";")
	return toks, node
}

func expr(toks []*Token) ([]*Token, Node) {
	return assign(toks)
}

func assign(toks []*Token) ([]*Token, Node) {
	toks, node := equality(toks)
	var isAssign bool
	var assignNode Node
	if toks, isAssign = consume(toks, "="); isAssign {
		toks, assignNode = assign(toks)
		node = NewAssignNode(node, assignNode)
	}
	return toks, node
}

func equality(toks []*Token) ([]*Token, Node) {
	toks, node := relational(toks)
	ops := []opKind{
		opKind{str: "==", kind: ndEq},
		opKind{str: "!=", kind: ndNeq},
	}
	matched := true
	for matched {
		matched = false
		for _, op := range ops {
			if newToks, isOp := consume(toks, op.str); isOp {
				nxtToks, relNode := relational(newToks)
				toks, node = nxtToks, NewArithNode(op.kind, node, relNode)
				matched = true
			}
		}
	}
	return toks, node
}

func relational(toks []*Token) ([]*Token, Node) {
	toks, node := addSub(toks)
	ops := []opKind{
		opKind{str: "<=", kind: ndLeq},
		opKind{str: ">=", kind: ndGeq},
		opKind{str: "<", kind: ndLt},
		opKind{str: ">", kind: ndGt},
	}
	matched := true
	for matched {
		matched = false
		for _, op := range ops {
			if newToks, isOp := consume(toks, op.str); isOp {
				nxtToks, addSubNode := addSub(newToks)
				toks, node = nxtToks, NewArithNode(op.kind, node, addSubNode)
				matched = true
			}
		}
	}
	return toks, node
}

func addSub(toks []*Token) ([]*Token, Node) {
	toks, node := mulDiv(toks)
	ops := []opKind{
		opKind{str: "+", kind: ndAdd},
		opKind{str: "-", kind: ndSub},
	}
	matched := true
	for matched {
		matched = false
		for _, op := range ops {
			var mulDivNode Node
			if newToks, isOp := consume(toks, op.str); isOp {
				toks, mulDivNode = mulDiv(newToks)
				node = NewArithNode(op.kind, node, mulDivNode)
				matched = true
			}
		}
	}
	return toks, node
}

func mulDiv(toks []*Token) ([]*Token, Node) {
	toks, node := unary(toks)
	ops := []opKind{
		opKind{str: "*", kind: ndMul},
		opKind{str: "/", kind: ndDiv},
	}
	matched := true
	for matched {
		matched = false
		for _, op := range ops {
			var unaryNode Node
			if newToks, isOp := consume(toks, op.str); isOp {
				toks, unaryNode = unary(newToks)
				node = NewArithNode(op.kind, node, unaryNode)
				matched = true
			}
		}
	}
	return toks, node
}

func unary(toks []*Token) ([]*Token, Node) {
	if newToks, isPlus := consume(toks, "+"); isPlus {
		return primary(newToks)
	} else if newToks, isMinus := consume(toks, "-"); isMinus {
		newToks, node := primary(newToks)
		return newToks, NewArithNode(ndSub, NewNumNode(0), node)
	}
	return primary(toks)
}

func primary(toks []*Token) ([]*Token, Node) {
	if toks, isPar := consume(toks, "("); isPar {
		toks, node := expr(toks)
		toks = expect(toks, ")")
		return toks, node
	}

	if toks, id, isID := consumeID(toks); isID {
		if toks, isLPar := consume(toks, "("); isLPar {
			var args []Node
			// no args
			if toks, isRPar := consume(toks, ")"); isRPar {
				return toks, NewFuncCallNode(id, args)
			}
			toks, arg := expr(toks)
			args = append(args, arg)
			for {
				newToks, isComma := consume(toks, ",")
				if !isComma {
					break
				}
				toks, arg = expr(newToks)
				args = append(args, arg)
			}
			toks = expect(toks, ")")
			return toks, NewFuncCallNode(id, args)
		}
		return toks, NewLvarNode(id)
	}

	toks, n := expectNum(toks)
	return toks, NewNumNode(n)
}
