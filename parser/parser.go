package parser

import (
	"Monkey/ast"
	"Monkey/lexer"
	"Monkey/token"
	"fmt"
	"strconv"
)

const (
	_ int = iota
	LOWEST
	EQUALS      // ==
	LESSGREATER // > or <
	SUM         // +
	PRODUCT     // *
	PREFIX      // -x or !x
	CALL        // myFunction()
	INDEX
)

var precedences = map[token.TokenType]int{
	token.ASSIGN:   EQUALS,
	token.EQ:       EQUALS,
	token.NOT_EQ:   EQUALS,
	token.LT:       LESSGREATER,
	token.GT:       LESSGREATER,
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.SLASH:    PRODUCT,
	token.ASTERISK: PRODUCT,
	token.LPAREN:   CALL,
	token.LBRACKET: INDEX,
}

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

type Parser struct {
	lex            *lexer.Lexer
	currToken      token.Token
	peekToken      token.Token
	errors         []string
	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

func New(lex *lexer.Lexer) *Parser {
	parser := &Parser{
		lex:    lex,
		errors: []string{},
	}

	// Prime the parser, read two tokens, so curToken and peekToken are both set
	parser.nextToken()
	parser.nextToken()

	// Prefix parse functions
	parser.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	parser.registerPrefix(token.IDENT, parser.parseIdentifier)
	parser.registerPrefix(token.INT, parser.parseIntegerLiteral)
	parser.registerPrefix(token.BANG, parser.parsePrefixExpression)
	parser.registerPrefix(token.MINUS, parser.parsePrefixExpression)
	parser.registerPrefix(token.TRUE, parser.parseBoolean)
	parser.registerPrefix(token.FALSE, parser.parseBoolean)
	parser.registerPrefix(token.LPAREN, parser.parseGroupedExpression)
	parser.registerPrefix(token.IF, parser.parseIfExpression)
	parser.registerPrefix(token.FUNCTION, parser.parseFunctionLiteral)
	parser.registerPrefix(token.STRING, parser.parseStringLiteral)
	parser.registerPrefix(token.LBRACKET, parser.parseArrayLiteral)
	parser.registerPrefix(token.LBRACE, parser.parseHashLiteral)

	// Infix parse functions
	parser.infixParseFns = make(map[token.TokenType]infixParseFn)
	parser.registerInfix(token.PLUS, parser.parseInfixExpression)
	parser.registerInfix(token.MINUS, parser.parseInfixExpression)
	parser.registerInfix(token.SLASH, parser.parseInfixExpression)
	parser.registerInfix(token.ASTERISK, parser.parseInfixExpression)
	parser.registerInfix(token.EQ, parser.parseInfixExpression)
	parser.registerInfix(token.ASSIGN, parser.parseAssignExpression)
	parser.registerInfix(token.NOT_EQ, parser.parseInfixExpression)
	parser.registerInfix(token.LT, parser.parseInfixExpression)
	parser.registerInfix(token.GT, parser.parseInfixExpression)
	parser.registerInfix(token.LPAREN, parser.parseCallExpression)
	parser.registerInfix(token.LBRACKET, parser.parseIndexExpression)

	return parser
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("Expected next token to be %s, but got %s instead", t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}

	p.peekError(t)
	return false
}

func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.currToken.Type == t
}

func (p *Parser) nextToken() {
	p.currToken = p.peekToken
	p.peekToken = p.lex.NextToken()
}

func (p *Parser) registerPrefix(token token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[token] = fn
}

func (p *Parser) registerInfix(token token.TokenType, fn infixParseFn) {
	p.infixParseFns[token] = fn
}

func (p *Parser) noPrefixParseFnError(t token.Token) {
	msg := fmt.Sprintf("no prefix parse function for token %s `%s` found", t.Type, t.Literal)
	p.errors = append(p.errors, msg)
}

func (p *Parser) peekPrecedence() int {
	if prec, ok := precedences[p.peekToken.Type]; ok {
		return prec

	}

	return LOWEST
}

func (p *Parser) currPrecedence() int {
	if prec, ok := precedences[p.currToken.Type]; ok {
		return prec
	}

	return LOWEST
}

func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()

		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}

		p.nextToken()
	}

	return program
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.currToken.Type {
	case token.LET:
		return p.parseLetStatement()

	case token.RETURN:
		return p.parseReturnStatement()

	default:
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseLetStatement() *ast.LetStatement {
	stmt := &ast.LetStatement{
		Token: p.currToken,
	}

	// expect next token would be `IDENTIFIER` and consume it
	if !p.expectPeek(token.IDENT) {
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.currToken, Value: p.currToken.Literal}

	// expect next token would be `ASSIGNMENT` and consume it
	if !p.expectPeek(token.ASSIGN) {
		return nil
	}

	p.nextToken() // Consume the `=` token

	stmt.Value = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken() // Handling semicolon since it is optional on repl
	}

	return stmt
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{
		Token: p.currToken,
	}

	p.nextToken()

	stmt.ReturnValue = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken() // Handling semicolon since it is optional on repl
	}

	return stmt
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{
		Token: p.currToken,
	}

	stmt.Expression = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	//      1 * 2 + 3
	//      ^   ^   ^
	//      |   |   |
	//      0   2   1
	//
	//
	//      1 + 2 * 3
	//      ^   ^   ^
	//      |   |   |
	//      0   1   2

	//      add(2,3)
	//       ^ ^
	//       | |
	//      _| |_
	//      |   |
	//  prefix infix

	prefix := p.prefixParseFns[p.currToken.Type]

	if prefix == nil {
		p.noPrefixParseFnError(p.currToken)
		return nil
	}

	leftExp := prefix()

	for !p.peekTokenIs(token.SEMICOLON) && p.peekPrecedence() > precedence {
		infix := p.infixParseFns[p.peekToken.Type]

		if infix == nil {
			return leftExp
		}

		// Advance the token so each parsing functions are inspecting `token.currToken`
		// Consume the infix operator token
		p.nextToken()

		leftExp = infix(leftExp)
	}

	return leftExp
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expr := &ast.InfixExpression{
		Token:    p.currToken,
		Operator: p.currToken.Literal,
		Left:     left,
	}

	precedence := p.currPrecedence() // Precedence of the infix operator

	p.nextToken()

	expr.Right = p.parseExpression(precedence)

	return expr
}

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.currToken, Value: p.currToken.Literal}
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	literal := &ast.IntegerLiteral{Token: p.currToken}

	value, err := strconv.ParseInt(p.currToken.Literal, 0, 64)

	if err != nil {
		msg := fmt.Sprintf("Could not parse %q as integer", p.currToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	literal.Value = value
	return literal
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	prefixExp := &ast.PrefixExpression{
		Token:    p.currToken,
		Operator: p.currToken.Literal,
	}

	p.nextToken()
	prefixExp.Right = p.parseExpression(PREFIX)

	return prefixExp

}

func (p *Parser) parseBoolean() ast.Expression {
	return &ast.Boolean{
		Token: p.currToken,
		Value: p.curTokenIs(token.TRUE),
	}
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	// If prefix parse function call to `parseExpression` it
	// have higher precedence and will parse the expression first
	p.nextToken() // consume the `(`

	expression := p.parseExpression(LOWEST)

	// 2 + (1 + 3) * 4
	if !p.expectPeek(token.RPAREN) { // consume the `)` cause `parseExpression` doesn't consume it
		return nil
	}

	return expression
}

func (p *Parser) parseIfExpression() ast.Expression {
	exp := &ast.IfExpression{Token: p.currToken}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	p.nextToken() // advance so `currToken` point to the expression after the `(`
	exp.Condition = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	if !p.expectPeek(token.LBRACE) { // advance to the `LBRACE` token
		return nil
	}

	exp.Consequence = p.parseBlockStatement()

	if p.peekTokenIs(token.ELSE) {
		p.nextToken() // consume the `else` token

		if !p.expectPeek(token.LBRACE) {
			return nil
		}

		exp.Alternative = p.parseBlockStatement()
	}

	return exp
}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.currToken}
	block.Statements = []ast.Statement{}

	p.nextToken() //consume the `{`

	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()

		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}

		p.nextToken()
	}

	return block
}

func (p *Parser) parseFunctionLiteral() ast.Expression {
	fun := &ast.FunctionLiteral{Token: p.currToken}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	fun.Parameters = p.parseFunctionParameters()

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	fun.Body = p.parseBlockStatement()

	return fun
}

func (p *Parser) parseFunctionParameters() []*ast.Identifier {
	idents := []*ast.Identifier{}

	// Empty function parameter list
	if p.peekTokenIs(token.RPAREN) {
		p.nextToken() // Consume the `)` token
		return idents
	}

	p.nextToken() // Advance the cursor, so we sit on first parameter list

	ident := &ast.Identifier{Token: p.currToken, Value: p.currToken.Literal}
	idents = append(idents, ident)

	for p.peekTokenIs(token.COMMA) {
		p.nextToken() // Eat the `,` token
		p.nextToken() // Advance the cursor so we sit on next parameter list

		ident = &ast.Identifier{Token: p.currToken, Value: p.currToken.Literal}
		idents = append(idents, ident)
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return idents
}

func (p *Parser) parseCallExpression(fn ast.Expression) ast.Expression {
	exp := &ast.CallExpression{Token: p.currToken, Function: fn}
	exp.Arguments = p.parseExpressionList(token.RPAREN)
	return exp
}

// This function is being replaced by `parseExpressionList`
func (p *Parser) parseCallArgument() []ast.Expression {
	args := []ast.Expression{}

	// Empty arguments
	if p.peekTokenIs(token.RPAREN) {
		p.nextToken() // Consume the `)` token
		return args
	}

	p.nextToken() // Advance the token so it sit on first argument list

	arg := p.parseExpression(LOWEST)
	args = append(args, arg)

	for p.peekTokenIs(token.COMMA) {
		p.nextToken() // Consume the `,` token
		p.nextToken() // Advance the cursor so it sit on next argument list

		arg = p.parseExpression(LOWEST)
		args = append(args, arg)
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}
	return args
}

func (p *Parser) parseStringLiteral() ast.Expression {
	str := &ast.StringLiteral{Token: p.currToken, Value: p.currToken.Literal}
	return str
}

func (p *Parser) parseArrayLiteral() ast.Expression {
	array := &ast.ArrayLiteral{Token: p.currToken}
	array.Elements = p.parseExpressionList(token.RBRACKET)
	return array
}

func (p *Parser) parseExpressionList(endToken token.TokenType) []ast.Expression {
	elements := []ast.Expression{}

	// either empty array `[]` or empty function call `fn()`
	if p.peekTokenIs(endToken) {
		p.nextToken() // Consume the closing pair.. either `]` or `)`
		return elements
	}

	p.nextToken() // Advance the cursor so it sit on the first expression

	elements = append(elements, p.parseExpression(LOWEST))

	for p.peekTokenIs(token.COMMA) {
		p.nextToken() // Consume the `,` token
		p.nextToken() // Advance the cursor so it sit on next expression

		elements = append(elements, p.parseExpression(LOWEST))
	}

	// Expect the closing pair.. either `[` or `)`
	if !p.expectPeek(endToken) {
		return nil
	}

	return elements

}

func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {

	ie := &ast.IndexExpression{Token: p.currToken, Left: left}

	p.nextToken() // Consume the `[` so we sit on the array index expression

	ie.Index = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RBRACKET) {
		return nil
	}

	return ie
}

func (p *Parser) parseAssignExpression(left ast.Expression) ast.Expression {
	ident := left.(*ast.Identifier)
	p.nextToken() // consume the `=` token
	return &ast.AssignmentExpression{Token: p.currToken, Name: ident, Value: p.parseExpression(LOWEST)}
}

func (p *Parser) parseHashLiteral() ast.Expression {
	hash := &ast.HashLiteral{Token: p.currToken}
	hash.Pairs = make(map[ast.Expression]ast.Expression)

	for !p.peekTokenIs(token.RBRACE) {
		p.nextToken() // Advance the cursor so we sit on the hash key expression

		hashKey := p.parseExpression(LOWEST)

		// Consume the `:` token
		if !p.expectPeek(token.COLON) {
			return nil
		}

		p.nextToken() // Advance the cursor so we sit on the hash val expression

		hashVal := p.parseExpression(LOWEST)

		hash.Pairs[hashKey] = hashVal

		if !p.peekTokenIs(token.RBRACE) && !p.expectPeek(token.COMMA) {
			return nil
		}

	}

	if !p.expectPeek(token.RBRACE) {
		return nil
	}

	return hash
}
