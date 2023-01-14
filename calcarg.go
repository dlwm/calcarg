package calcarg

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
)

// 词元
const (
	eof    = "EOF"    // 结束符
	digit  = "DIGIT"  // 数字
	letter = "LETTER" // 字母

	dot = "." // 小数点

	plus     = "+"
	minus    = "-"
	asterisk = "*"
	slash    = "/"

	lparen = "("
	rparen = ")"

	lEscape = "<"
	rEscape = ">"
)

// 优先级
const (
	_ int = iota
	lowest
	sum     // +, -
	product // *, /
	prefix  // -X
	call    // (X)
	escape  // <I>
)

// 优先级对应表
var precedences = map[string]int{
	plus:     sum,
	minus:    sum,
	slash:    product,
	asterisk: product,
	lparen:   call,
	lEscape:  escape,
}

// Calculator 计算控制器
type Calculator struct {
	Formula string
	Root    expression
}

// Analyse 解析公式
func Analyse(formula string) (*Calculator, error) {
	lexer := newLex(formula)
	parser := newParser(lexer)

	exp := parser.parseExpression(lowest)
	calculator := &Calculator{
		Formula: formula,
		Root:    exp,
	}

	return calculator, nil
}

//// Eval 根据参数计算
//func (c *Calculator) Eval(argsJson string) (int64, error) {
//	args := make(map[string]interface{})
//	err := json.Unmarshal([]byte(argsJson), &args)
//	if err != nil {
//		return 0, err
//	}
//
//	return eval(c.Root, args)
//}

// Eval 根据参数计算
func (c *Calculator) Eval(args map[string]float32) (float32, error) {
	return eval(c.Root, args)
}

// 表达式计算
func eval(exp expression, args map[string]float32) (float32, error) {
	switch node := exp.(type) {
	case *integerLiteralExpression:
		return node.Value, nil
	case *letterLiteralExpression:
		value, ok := args[node.Key]
		if !ok {
			return 0, errors.New(fmt.Sprintf("Cannot find '%s' in args", node.Key))
		}

		return value, nil
	case *prefixExpression:
		rightV, err := eval(node.Right, args)
		if err != nil {
			return 0, err
		}
		return evalPrefixExpression(node.Operator, rightV), nil
	case *infixExpression:
		leftV, err := eval(node.Left, args)
		if err != nil {
			return 0, err
		}
		rightV, err := eval(node.Right, args)
		if err != nil {
			return 0, err
		}
		return evalInfixExpression(leftV, node.Operator, rightV), nil
	}

	return 0, nil
}

// 计算前缀表达式
func evalPrefixExpression(operator string, right float32) float32 {
	if operator != "-" {
		return 0
	}
	return -right
}

// 计算双目表达式
func evalInfixExpression(left float32, operator string, right float32) float32 {
	switch operator {
	case "+":
		return left + right
	case "-":
		return left - right
	case "*":
		return left * right
	case "/":
		if right != 0 {
			return left / right
		} else {
			return 0
		}
	default:
		return 0
	}
}

type token struct {
	Type    string
	Literal string
}

func newToken(tokenType string, c byte) token {
	return token{
		Type:    tokenType,
		Literal: string(c),
	}
}

type lexer struct {
	input        string
	position     int
	readPosition int
	ch           byte
}

func newLex(input string) *lexer {
	l := &lexer{input: input}
	l.readChar()
	return l
}

func (l *lexer) nextToken() token {
	var tok token

	l.skipWhitespace()

	switch l.ch {
	case '<':
		tok = newToken(lEscape, l.ch)
	case '>':
		tok = newToken(rEscape, l.ch)
	case '(':
		tok = newToken(lparen, l.ch)
	case ')':
		tok = newToken(rparen, l.ch)
	case '+':
		tok = newToken(plus, l.ch)
	case '-':
		tok = newToken(minus, l.ch)
	case '/':
		tok = newToken(slash, l.ch)
	case '*':
		tok = newToken(asterisk, l.ch)
	case 0:
		tok.Literal = ""
		tok.Type = eof
	default:
		if isDigit(l.ch) {
			tok.Type = digit
			tok.Literal = l.readNumber()
			return tok
		} else if isLetter(l.ch) {
			tok.Type = letter
			tok.Literal = l.readWord()
			return tok
		} else {
			panic(fmt.Sprintf("invalid char: '%c'", l.ch))
		}
	}

	l.readChar()
	return tok
}

func isDot(ch byte) bool {
	return '.' == ch
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

func (l *lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition += 1
}

func (l *lexer) readNumber() string {
	HadDot := false
	position := l.position
	for isDigit(l.ch) || isDot(l.ch) {
		if isDot(l.ch) {
			if HadDot {
				panic(fmt.Sprintf("incorrect dot use: '%c'", l.ch))
			}
			HadDot = true
		}
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *lexer) readWord() string {
	position := l.position
	for isDigit(l.ch) || isLetter(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

type expression interface {
	string() string
}

type integerLiteralExpression struct {
	Token token
	Value float32
}

func (il *integerLiteralExpression) string() string { return il.Token.Literal }

type letterLiteralExpression struct {
	Token token
	Key   string
}

func (ll *letterLiteralExpression) string() string { return ll.Token.Literal }

type prefixExpression struct {
	Token    token
	Operator string
	Right    expression
}

func (pe *prefixExpression) string() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(pe.Operator)
	out.WriteString(pe.Right.string())
	out.WriteString(")")

	return out.String()
}

type infixExpression struct {
	Token    token
	Left     expression
	Operator string
	Right    expression
}

func (ie *infixExpression) string() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(ie.Left.string())
	out.WriteString(" ")
	out.WriteString(ie.Operator)
	out.WriteString(" ")
	out.WriteString(ie.Right.string())
	out.WriteString(")")

	return out.String()
}

// parser
type (
	prefixParseFn func() expression
	infixParseFn  func(expression) expression
)

type parser struct {
	l *lexer

	curToken  token
	peekToken token

	prefixParseFns map[string]prefixParseFn
	infixParseFns  map[string]infixParseFn

	errors []string
}

func (p *parser) registerPrefix(tokenType string, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *parser) registerInfix(tokenType string, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

func newParser(l *lexer) *parser {
	p := &parser{
		l:      l,
		errors: []string{},
	}

	p.prefixParseFns = make(map[string]prefixParseFn)
	p.registerPrefix(digit, p.parseIntegerLiteral)
	p.registerPrefix(letter, p.parseLetterLiteral)
	p.registerPrefix(minus, p.parsePrefixExpression)
	p.registerPrefix(lparen, p.parseGroupedExpression)
	p.registerPrefix(lEscape, p.parseEscapedExpression)

	p.infixParseFns = make(map[string]infixParseFn)
	p.registerInfix(plus, p.parseInfixExpression)
	p.registerInfix(minus, p.parseInfixExpression)
	p.registerInfix(slash, p.parseInfixExpression)
	p.registerInfix(asterisk, p.parseInfixExpression)

	p.nextToken()
	p.nextToken()

	return p
}

func (p *parser) parseExpression(precedence int) expression {
	prefix := p.prefixParseFns[p.curToken.Type]
	returnExp := prefix()

	for precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return returnExp
		}

		p.nextToken()
		returnExp = infix(returnExp)
	}

	return returnExp
}

func (p *parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return lowest
}

func (p *parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}
	return lowest
}

func (p *parser) peekError(t string) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instend",
		t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

func (p *parser) expectPeek(t string) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	} else {
		p.peekError(t)
		return false
	}
}

func (p *parser) peekTokenIs(t string) bool {
	return p.peekToken.Type == t
}

func (p *parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.nextToken()
}

func (p *parser) parseIntegerLiteral() expression {
	lit := &integerLiteralExpression{Token: p.curToken}

	value, err := strconv.ParseFloat(p.curToken.Literal, 32)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	lit.Value = float32(value)
	return lit
}

func (p *parser) parseLetterLiteral() expression {
	lit := &letterLiteralExpression{
		Token: p.curToken,
		Key:   p.curToken.Literal,
	}

	return lit
}

func (p *parser) parsePrefixExpression() expression {
	expression := &prefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}
	p.nextToken()
	expression.Right = p.parseExpression(prefix)
	return expression
}

func (p *parser) parseGroupedExpression() expression {
	p.nextToken()
	exp := p.parseExpression(lowest)

	if !p.expectPeek(rparen) {
		return nil
	}
	return exp
}

func (p *parser) parseEscapedExpression() expression {
	p.nextToken()
	exp := p.parseExpression(lowest)

	if !p.expectPeek(rEscape) {
		return nil
	}
	return exp
}

func (p *parser) parseInfixExpression(left expression) expression {
	expression := &infixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}

	precedence := p.curPrecedence()
	p.nextToken()

	// // 通过降低优先级，来达到右结合
	//if expression.Operator == "+" {
	//	expression.Right = p.parseExpression(precedence - 1)
	//} else {
	//	expression.Right = p.parseExpression(precedence)
	//}
	expression.Right = p.parseExpression(precedence)

	return expression
}
