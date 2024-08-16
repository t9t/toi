package main

import (
	"fmt"
	"strconv"
)

type Parser struct {
	tokens []Token

	loopBodyCount int
	forCounter    int
}

func (p *Parser) consume(i int) {
	p.tokens = p.tokens[i:]
}

func (p *Parser) hasCurrent() bool {
	return len(p.tokens) != 0
}

func (p *Parser) current() Token {
	return p.tokens[0]
}

func (p *Parser) left() int {
	return len(p.tokens)
}

func (p *Parser) hasNext() bool {
	return len(p.tokens) >= 2
}

func (p *Parser) next() Token {
	return p.nextN(1)
}

func (p *Parser) nextN(n int) Token {
	return p.tokens[n]
}

func (p *Parser) eof() bool {
	return len(p.tokens) == 0
}

func (p *Parser) parse() (Statement, error) {
	statements := make([]Statement, 0)
	for !p.eof() {
		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		if stmt != nil {
			statements = append(statements, stmt)
		}
	}

	return &BlockStatement{Statements: statements}, nil
}

func (p *Parser) parseStatement() (stmt Statement, err error) {
	if p.current().Type == TokenNewline {
		// Skip over empty lines
		p.consume(1)
		return nil, nil
	}

	if p.current().Type == TokenIf {
		stmt, err = p.parseIfStatement()
		if err != nil {
			return nil, err
		}
	} else if p.current().Type == TokenWhile {
		stmt, err = p.parseWhileStatement()
		if err != nil {
			return nil, err
		}
	} else if p.current().Type == TokenFor {
		stmt, err = p.parseForStatement()
		if err != nil {
			return nil, err
		}
	} else if p.current().Type == TokenExit {
		stmt, err = p.parseExitLoopStatement()
		if err != nil {
			return nil, err
		}
	} else if p.current().Type == TokenNext {
		stmt, err = p.parseNextIterationStatement()
		if err != nil {
			return nil, err
		}
	} else {
		stmt, err = p.parseAssignmentStatement()
		if err != nil {
			return nil, err
		}
	}

	if p.hasCurrent() && p.current().Type != TokenNewline && p.current().Type != TokenBraceClose {
		tok := p.current()
		return nil, fmt.Errorf("expected newline after statement but got %s ('%s') at %d:%d", tok.Type, tok.Lexeme, tok.Line, tok.Col)
	}

	if p.hasCurrent() && p.current().Type == TokenNewline {
		// Consume newline
		p.consume(1)
	}
	return stmt, nil
}

func (p *Parser) parseBlock(typ string) (Statement, error) {
	if !p.hasCurrent() || p.current().Type != TokenBraceOpen {
		tok := p.current()
		return nil, fmt.Errorf("expected '{' after %s at %d:%d", typ, tok.Line, tok.Col)
	}

	token := p.current()

	p.consume(1)
	statements := make([]Statement, 0)
	for p.hasCurrent() && p.current().Type != TokenBraceClose {
		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		} else if stmt == nil {
			continue
		}
		statements = append(statements, stmt)
	}

	if !p.hasCurrent() || p.current().Type != TokenBraceClose {
		tok := p.current()
		return nil, fmt.Errorf("expected '}' after %s statements at %d:%d", typ, tok.Line, tok.Col)
	}

	p.consume(1)
	return &BlockStatement{Token: token, Statements: statements}, nil
}

func (p *Parser) parseIfStatement() (Statement, error) {
	token := p.current()

	p.consume(1)
	expr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	block, err := p.parseBlock("if expression")
	if err != nil {
		return nil, err
	}

	var otherwiseBlock *Statement
	if p.hasCurrent() && p.current().Type == TokenOtherwise {
		p.consume(1)
		otherwise, err := p.parseBlock("otherwise")
		if err != nil {
			return nil, err
		}
		otherwiseBlock = &otherwise
	}

	return &IfStatement{Token: token, Condition: expr, Then: block, Otherwise: otherwiseBlock}, nil
}

func (p *Parser) parseWhileStatement() (Statement, error) {
	token := p.current()
	p.consume(1)

	expr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	p.loopBodyCount += 1
	block, err := p.parseBlock("while expression")
	if err != nil {
		return nil, err
	}
	p.loopBodyCount -= 1

	return &WhileStatement{Token: token, Condition: expr, Body: block}, nil
}

func (p *Parser) parseForStatement() (Statement, error) {
	// for value = [arrayOrMap]indexOrKey { ... }
	token := p.current()

	if p.left() < 4 {
		return nil, fmt.Errorf("incomplete 'for' statement at %d:%d", token.Line, token.Col)
	}

	p.consume(1)
	if p.current().Type != TokenIdentifier {
		tok := p.current()
		return nil, fmt.Errorf("expected identifier after 'for' but got '%v' at %d:%d", tok.Type, tok.Line, tok.Col)
	} else if p.next().Type != TokenEquals {
		tok := p.next()
		return nil, fmt.Errorf("expected '=' after 'for' identifier but got '%v' at %d:%d", tok.Type, tok.Line, tok.Col)
	} else if p.nextN(2).Type != TokenBracketOpen {
		tok := p.nextN(2)
		return nil, fmt.Errorf("expected '[' after '=' in 'for' but got '%v' at %d:%d", tok.Type, tok.Line, tok.Col)
	}

	valueIdentifier := p.current()
	p.consume(3) // identifier, equals, and bracket open

	containerExpression, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	if p.left() < 2 || p.current().Type != TokenBracketClose || p.next().Type != TokenIdentifier {
		tok := p.current()
		return nil, fmt.Errorf("expected ']' and index identifier 'for' container expression but got '%v' at %d:%d", tok.Type, tok.Line, tok.Col)
	}

	keyIdentifier := p.next()

	p.consume(2)

	p.loopBodyCount += 1
	block, err := p.parseBlock("for expression")
	if err != nil {
		return nil, err
	}
	p.loopBodyCount -= 1

	ident := func(s string) Token { return Token{Type: TokenIdentifier, Lexeme: s} }

	p.forCounter += 1
	f := strconv.Itoa(p.forCounter)
	containerIdent := ident("_for_container_" + f)
	containerExpr := &VariableExpression{containerIdent}
	keysIdent := ident("_for_keys_" + f)
	keysExpr := &VariableExpression{keysIdent}
	indexIdent := ident("_for_index_" + f)
	indexExpr := &VariableExpression{indexIdent}

	return &BlockStatement{
		Token: token,
		Statements: []Statement{
			&AssignmentStatement{ // _for_container = (container expression)
				Identifier: containerIdent,
				Expression: containerExpression,
			},
			&AssignmentStatement{ // _for_keys = keys(_for_container)
				Identifier: keysIdent,
				Expression: &FunctionCallExpression{
					Token:        keysIdent,
					FunctionName: "keys",
					Arguments:    []Expression{containerExpr},
				},
			},
			&AssignmentStatement{ // _for_index = 0
				Identifier: indexIdent,
				Expression: &LiteralExpression{Token{Type: TokenNumber, Lexeme: "0", Literal: 0}},
			}, // _i
			&WhileStatement{
				Token: token,
				Condition: &BinaryExpression{ // _for_index < len(_for_keys)
					Left:     indexExpr,
					Operator: Token{Type: TokenLessThan, Lexeme: "<"},
					Right: &FunctionCallExpression{
						Token:        Token{Type: TokenIdentifier, Lexeme: "len"},
						FunctionName: "len",
						Arguments:    []Expression{keysExpr},
					},
				},
				Body: &BlockStatement{
					Token: token,
					Statements: []Statement{
						&AssignmentStatement{ // (used defined) key = [_for_keys]_for_index
							Identifier: keyIdentifier,
							Expression: &FunctionCallExpression{ // get(_for_keys, _for_index)
								Token:        Token{Type: TokenIdentifier, Lexeme: "get"},
								FunctionName: "get",
								Arguments:    []Expression{keysExpr, indexExpr},
							},
						},
						&AssignmentStatement{ // (user defined) value = [container]key
							Identifier: valueIdentifier,
							Expression: &FunctionCallExpression{ // get(container, key)
								Token:        Token{Type: TokenIdentifier, Lexeme: "get"},
								FunctionName: "get",
								Arguments:    []Expression{containerExpr, &VariableExpression{keyIdentifier}},
							},
						},
						block,
					},
				},
				AfterBody: &AssignmentStatement{ // _for_index = _for_index + 1
					Identifier: indexIdent,
					Expression: &BinaryExpression{
						Left:     indexExpr,
						Operator: Token{Type: TokenPlus, Lexeme: "+"},
						Right:    &LiteralExpression{Token{Type: TokenNumber, Lexeme: "1", Literal: 1}},
					},
				},
			},
		},
	}, nil
}

func (p *Parser) parseExitLoopStatement() (Statement, error) {
	token := p.current()

	if !p.hasNext() || p.next().Type != TokenLoop {
		tok := p.next()
		return nil, fmt.Errorf("expected 'loop' after 'exit' at %d:%d", tok.Line, tok.Col)
	}

	if p.loopBodyCount == 0 {
		tok := token
		return nil, fmt.Errorf("can only use 'exit loop' in 'while' or 'for' body at %d:%d", tok.Line, tok.Col)
	}

	p.consume(2)
	return &ExitLoopStatement{Token: token}, nil
}

func (p *Parser) parseNextIterationStatement() (Statement, error) {
	token := p.current()

	if !p.hasNext() || p.next().Type != TokenIteration {
		tok := p.next()
		return nil, fmt.Errorf("expected 'iteration' after 'next' at %d:%d", tok.Line, tok.Col)
	}

	if p.loopBodyCount == 0 {
		tok := token
		return nil, fmt.Errorf("can only use 'next iteration' in 'while' or 'for' body at %d:%d", tok.Line, tok.Col)
	}

	p.consume(2)
	return &NextIterationStatement{Token: token}, nil
}

func (p *Parser) parseAssignmentStatement() (Statement, error) {
	startToken := p.current()
	left, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	if !p.hasCurrent() || p.current().Type != TokenEquals {
		return &ExpressionStatement{startToken, left}, nil
	}
	p.consume(1)

	right, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	if access, ok := left.(*ContainerAccessExpression); ok {
		return &ExpressionStatement{
			Token: startToken,
			Expression: &FunctionCallExpression{
				Token:        access.Token,
				FunctionName: "set",
				Arguments:    []Expression{access.Container, access.Access, right},
			},
		}, nil
	}

	return &AssignmentStatement{Identifier: startToken, Expression: right}, nil
}

func (p *Parser) parseExpression() (Expression, error) {
	return p.parseLogicalOr()
}

func (p *Parser) parseContainerAccess() (Expression, error) {
	startToken := p.current()
	nestedLevel := 0
	for p.hasCurrent() && p.current().Type == TokenBracketOpen {
		nestedLevel += 1
		p.consume(1)
	}

	innerExpression, err := p.parsePrimary()
	if err != nil {
		return nil, err
	} else if nestedLevel == 0 {
		return innerExpression, nil
	}

	for i := 0; i < nestedLevel; i += 1 {
		if p.current().Type != TokenBracketClose {
			tok := p.current()
			return nil, fmt.Errorf("expected ']' after '[' and expression but got '%v' at %d:%d", tok.Type, tok.Line, tok.Col)
		}
		p.consume(1)

		indexExpr, err := p.parsePrimary()
		if err != nil {
			return nil, err
		}

		innerExpression = &ContainerAccessExpression{Token: startToken, Container: innerExpression, Access: indexExpr}
	}
	return innerExpression, nil
}

func (p *Parser) parseLogicalOr() (Expression, error) {
	return p.parseBinary(TokenPipe, p.parseLogicalAnd)
}

func (p *Parser) parseLogicalAnd() (Expression, error) {
	return p.parseBinary(TokenAmpersand, p.parseEqualEqual)
}

func (p *Parser) parseEqualEqual() (Expression, error) {
	return p.parseBinary(TokenEqualEqual, p.parseNotEqual)
}

func (p *Parser) parseNotEqual() (Expression, error) {
	return p.parseBinary(TokenNotEqual, p.parseGreaterThan)
}

func (p *Parser) parseGreaterThan() (Expression, error) {
	return p.parseBinary(TokenGreaterThan, p.parseGreaterEqual)
}

func (p *Parser) parseGreaterEqual() (Expression, error) {
	return p.parseBinary(TokenGreaterEqual, p.parseLessThan)
}

func (p *Parser) parseLessThan() (Expression, error) {
	return p.parseBinary(TokenLessThan, p.parseLessEqual)
}

func (p *Parser) parseLessEqual() (Expression, error) {
	return p.parseBinary(TokenLessEqual, p.parseMinus)
}

func (p *Parser) parseMinus() (Expression, error) {
	return p.parseBinary(TokenMinus, p.parsePlus)
}

func (p *Parser) parsePlus() (Expression, error) {
	return p.parseBinary(TokenPlus, p.parseDivide)
}

func (p *Parser) parseDivide() (Expression, error) {
	return p.parseBinary(TokenSlash, p.parseMultiply)
}

func (p *Parser) parseMultiply() (Expression, error) {
	return p.parseBinary(TokenAsterisk, p.parseStringConcat)
}

func (p *Parser) parseStringConcat() (Expression, error) {
	return p.parseBinary(TokenUnderscore, p.parseContainerAccess)
}

func (p *Parser) parseBinary(tokenType TokenType, down func() (Expression, error)) (Expression, error) {
	left, err := down()
	if err != nil {
		return nil, err
	}

	for p.hasCurrent() && p.current().Type == tokenType {
		operator := p.current()
		p.consume(1)
		right, err := down()
		if err != nil {
			return nil, err
		}

		left = &BinaryExpression{Left: left, Operator: operator, Right: right}
	}

	return left, nil
}

func (p *Parser) parsePrimary() (Expression, error) {
	if !p.hasCurrent() {
		return nil, fmt.Errorf("expected primary expression but reached end of data")
	}

	token := p.current()
	if token.Type == TokenString || token.Type == TokenNumber {
		p.consume(1)
		return &LiteralExpression{Token: token}, nil
	} else if token.Type == TokenIdentifier {
		if p.left() >= 2 && p.next().Type == TokenParenOpen {
			return p.parseFunctionCall()
		}

		// Variable access
		p.consume(1)
		return &VariableExpression{Token: token}, nil
	} else if token.Type == TokenParenOpen {
		p.consume(1)
		expr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}

		if !p.hasCurrent() || p.current().Type != TokenParenClose {
			tok := p.current()
			return nil, fmt.Errorf("expect ')' after '(' and expression but got '%v' at %d:%d", tok.Type, tok.Line, tok.Col)
		}

		p.consume(1)
		return expr, nil
	}

	return nil, fmt.Errorf("expected primary expression but got %s ('%s') at %d:%d", token.Type, token.Lexeme, token.Line, token.Col)
}

func (p *Parser) parseFunctionCall() (Expression, error) {
	callToken := p.current()
	identifier := callToken.Lexeme

	builtin, found := builtins[identifier]
	if !found {
		tok := callToken
		return nil, fmt.Errorf("no such builtin function '%s' at %d:%d", identifier, tok.Line, tok.Col)
	}

	p.consume(2) // Consume identifier and '('

	arguments := make([]Expression, 0)
	for p.hasCurrent() {
		// TODO: remove duplication
		if p.current().Type == TokenParenClose {
			p.consume(1)
			break
		}

		expr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}

		arguments = append(arguments, expr)
		if p.hasCurrent() {
			if p.current().Type == TokenComma {
				p.consume(1)
			} else if p.current().Type != TokenParenClose {
				tok := p.current()
				return nil, fmt.Errorf("expected ')' or ',' but got %s ('%s') at %d:%d", tok.Type, tok.Lexeme, tok.Line, tok.Col)
			}
		} else {
			return nil, fmt.Errorf("expected ')' or ',' but got end of input")
		}
	}

	if len(arguments) != builtin.Arity && builtin.Arity != ArityVariadic {
		tok := p.current()
		return nil, fmt.Errorf("expected %d arguments but got %d for function '%s' at %d:%d", builtin.Arity, len(arguments), identifier, tok.Line, tok.Col)
	}

	return &FunctionCallExpression{Token: callToken, FunctionName: callToken.Lexeme, Arguments: arguments}, nil
}
