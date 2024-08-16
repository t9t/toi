package main

import (
	"fmt"
	"strconv"
)

// TODO: global state is yuck, don't do it
var loopBodyCount = 0
var forCounter = 0

type Parser struct {
	tokens []Token
	index  int
}

func (p *Parser) previous() Token {
	return p.tokens[p.index-1]
}

func (p *Parser) current() Token {
	return p.tokens[p.index]
}

func (p *Parser) hasNext() bool {
	return p.index < len(p.tokens)-1
}

func (p *Parser) next() Token {
	return p.tokens[p.index+1]
}

func (p *Parser) eof() bool {
	return p.index == len(p.tokens)
}

func (p *Parser) parse(tokens []Token) (Statement, error) {
	statements := make([]Statement, 0)
	for len(tokens) > 0 {
		stmt, next, err := p.parseStatement(tokens)
		if err != nil {
			return nil, err
		}
		if stmt != nil {
			statements = append(statements, stmt)
		}
		tokens = next
	}

	return &BlockStatement{Statements: statements}, nil
}

func (p *Parser) parseStatement(tokens []Token) (stmt Statement, next []Token, err error) {
	if tokens[0].Type == TokenNewline {
		// Skip over empty lines
		return nil, tokens[1:], nil
	}

	if tokens[0].Type == TokenIf {
		stmt, next, err = p.parseIfStatement(tokens)
		if err != nil {
			return nil, next, err
		}
	} else if tokens[0].Type == TokenWhile {
		stmt, next, err = p.parseWhileStatement(tokens)
		if err != nil {
			return nil, next, err
		}
	} else if tokens[0].Type == TokenFor {
		stmt, next, err = p.parseForStatement(tokens)
		if err != nil {
			return nil, next, err
		}
	} else if tokens[0].Type == TokenExit {
		stmt, next, err = p.parseExitLoopStatement(tokens)
		if err != nil {
			return nil, next, err
		}
	} else if tokens[0].Type == TokenNext {
		stmt, next, err = p.parseNextIterationStatement(tokens)
		if err != nil {
			return nil, next, err
		}
	} else {
		stmt, next, err = p.parseAssignmentStatement(tokens)
		if err != nil {
			return nil, next, err
		}
	}

	if len(next) != 0 && next[0].Type != TokenNewline && next[0].Type != TokenBraceClose {
		tok := next[0]
		return nil, nil, fmt.Errorf("expected newline after statement but got %s ('%s') at %d:%d", tok.Type, tok.Lexeme, tok.Line, tok.Col)
	}

	if len(next) != 0 && next[0].Type == TokenNewline {
		// Consume newline
		next = next[1:]
	}
	return stmt, next, nil
}

func (p *Parser) parseBlock(tokens []Token, typ string) (Statement, []Token, error) {
	next := tokens

	if len(next) == 0 || next[0].Type != TokenBraceOpen {
		tok := next[0]
		return nil, nil, fmt.Errorf("expected '{' after %s at %d:%d", typ, tok.Line, tok.Col)
	}

	token := next[0]

	next = next[1:]
	statements := make([]Statement, 0)
	for len(next) != 0 && next[0].Type != TokenBraceClose {
		stmt, next2, err := p.parseStatement(next)
		if err != nil {
			return nil, next2, err
		} else if stmt == nil {
			next = next2
			continue
		}
		statements = append(statements, stmt)
		next = next2
	}

	if len(next) == 0 || next[0].Type != TokenBraceClose {
		tok := next[0]
		return nil, nil, fmt.Errorf("expected '}' after %s statements at %d:%d", typ, tok.Line, tok.Col)
	}

	return &BlockStatement{Token: token, Statements: statements}, next[1:], nil
}

func (p *Parser) parseIfStatement(tokens []Token) (Statement, []Token, error) {
	token := tokens[0]

	expr, next, err := p.parseExpression(tokens[1:])
	if err != nil {
		return nil, nil, err
	}

	block, next, err := p.parseBlock(next, "if expression")
	if err != nil {
		return nil, nil, err
	}

	var otherwiseBlock *Statement
	if len(next) > 0 && next[0].Type == TokenOtherwise {
		next = next[1:] // Consume "otherwise"
		otherwise, otherwiseNext, err := p.parseBlock(next, "otherwise")
		if err != nil {
			return nil, nil, err
		}
		next = otherwiseNext
		otherwiseBlock = &otherwise
	}

	return &IfStatement{Token: token, Condition: expr, Then: block, Otherwise: otherwiseBlock}, next, nil
}

func (p *Parser) parseWhileStatement(tokens []Token) (Statement, []Token, error) {
	token := tokens[0]

	expr, next, err := p.parseExpression(tokens[1:])
	if err != nil {
		return nil, nil, err
	}

	loopBodyCount += 1
	block, next, err := p.parseBlock(next, "while expression")
	if err != nil {
		return nil, nil, err
	}
	loopBodyCount -= 1

	return &WhileStatement{Token: token, Condition: expr, Body: block}, next, nil
}

func (p *Parser) parseForStatement(tokens []Token) (Statement, []Token, error) {
	// for value = [arrayOrMap]indexOrKey { ... }
	token := tokens[0]

	if len(tokens) < 4 {
		return nil, nil, fmt.Errorf("incomplete 'for' statement at %d:%d", token.Line, token.Col)
	}

	if tokens[1].Type != TokenIdentifier {
		tok := tokens[1]
		return nil, nil, fmt.Errorf("expected identifier after 'for' but got '%v' at %d:%d", tok.Type, tok.Line, tok.Col)
	} else if tokens[2].Type != TokenEquals {
		tok := tokens[2]
		return nil, nil, fmt.Errorf("expected '=' after 'for' identifier but got '%v' at %d:%d", tok.Type, tok.Line, tok.Col)
	} else if tokens[3].Type != TokenBracketOpen {
		tok := tokens[3]
		return nil, nil, fmt.Errorf("expected '[' after '=' in 'for' but got '%v' at %d:%d", tok.Type, tok.Line, tok.Col)
	}

	valueIdentifier := tokens[1]

	containerExpression, next, err := p.parseExpression(tokens[4:])
	if err != nil {
		return nil, nil, err
	}

	if len(next) < 2 || next[0].Type != TokenBracketClose || next[1].Type != TokenIdentifier {
		tok := next[0]
		return nil, nil, fmt.Errorf("expected ']' and index identifier 'for' container expression but got '%v' at %d:%d", tok.Type, tok.Line, tok.Col)
	}

	keyIdentifier := next[1]

	next = next[2:]

	loopBodyCount += 1
	block, next, err := p.parseBlock(next, "for expression")
	if err != nil {
		return nil, nil, err
	}
	loopBodyCount -= 1

	ident := func(s string) Token { return Token{Type: TokenIdentifier, Lexeme: s} }

	forCounter += 1
	f := strconv.Itoa(forCounter)
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
	}, next, nil
}

func (p *Parser) parseExitLoopStatement(tokens []Token) (Statement, []Token, error) {
	token := tokens[0]

	if len(tokens) == 1 || tokens[1].Type != TokenLoop {
		tok := tokens[1]
		return nil, nil, fmt.Errorf("expected 'loop' after 'exit' at %d:%d", tok.Line, tok.Col)
	}

	if loopBodyCount == 0 {
		tok := tokens[0]
		return nil, nil, fmt.Errorf("can only use 'exit loop' in 'while' or 'for' body at %d:%d", tok.Line, tok.Col)
	}

	return &ExitLoopStatement{Token: token}, tokens[2:], nil
}

func (p *Parser) parseNextIterationStatement(tokens []Token) (Statement, []Token, error) {
	token := tokens[0]

	if len(tokens) == 1 || tokens[1].Type != TokenIteration {
		tok := tokens[1]
		return nil, nil, fmt.Errorf("expected 'iteration' after 'next' at %d:%d", tok.Line, tok.Col)
	}

	if loopBodyCount == 0 {
		tok := tokens[0]
		return nil, nil, fmt.Errorf("can only use 'next iteration' in 'while' or 'for' body at %d:%d", tok.Line, tok.Col)
	}

	return &NextIterationStatement{Token: token}, tokens[2:], nil
}

func (p *Parser) parseAssignmentStatement(tokens []Token) (Statement, []Token, error) {
	startToken := tokens[0]
	left, next, err := p.parseExpression(tokens)
	if err != nil {
		return nil, nil, err
	}

	if len(next) == 0 || next[0].Type != TokenEquals {
		return &ExpressionStatement{startToken, left}, next, nil
	}
	next = next[1:]

	right, next, err := p.parseExpression(next)
	if err != nil {
		return nil, nil, err
	}

	if access, ok := left.(*ContainerAccessExpression); ok {
		return &ExpressionStatement{
			Token: startToken,
			Expression: &FunctionCallExpression{
				Token:        access.Token,
				FunctionName: "set",
				Arguments:    []Expression{access.Container, access.Access, right},
			},
		}, next, nil
	}

	return &AssignmentStatement{Identifier: tokens[0], Expression: right}, next, nil
}

func (p *Parser) parseExpression(tokens []Token) (Expression, []Token, error) {
	return p.parseLogicalOr(tokens)
}

func (p *Parser) parseContainerAccess(tokens []Token) (Expression, []Token, error) {
	startToken := tokens[0]
	nestedLevel := 0
	for len(tokens) != 0 && tokens[0].Type == TokenBracketOpen {
		nestedLevel += 1
		tokens = tokens[1:]
	}

	innerExpression, next, err := p.parsePrimary(tokens)
	if err != nil {
		return nil, nil, err
	} else if nestedLevel == 0 {
		return innerExpression, next, nil
	}

	for i := 0; i < nestedLevel; i += 1 {
		if next[0].Type != TokenBracketClose {
			tok := next[0]
			return nil, nil, fmt.Errorf("expected ']' after '[' and expression but got '%v' at %d:%d", tok.Type, tok.Line, tok.Col)
		}
		next = next[1:]

		indexExpr, more, err := p.parsePrimary(next)
		if err != nil {
			return nil, nil, err
		}

		innerExpression = &ContainerAccessExpression{Token: startToken, Container: innerExpression, Access: indexExpr}
		next = more
	}
	return innerExpression, next, nil
}

func (p *Parser) parseLogicalOr(tokens []Token) (Expression, []Token, error) {
	return p.parseBinary(tokens, TokenPipe, p.parseLogicalAnd)
}

func (p *Parser) parseLogicalAnd(tokens []Token) (Expression, []Token, error) {
	return p.parseBinary(tokens, TokenAmpersand, p.parseEqualEqual)
}

func (p *Parser) parseEqualEqual(tokens []Token) (Expression, []Token, error) {
	return p.parseBinary(tokens, TokenEqualEqual, p.parseNotEqual)
}

func (p *Parser) parseNotEqual(tokens []Token) (Expression, []Token, error) {
	return p.parseBinary(tokens, TokenNotEqual, p.parseGreaterThan)
}

func (p *Parser) parseGreaterThan(tokens []Token) (Expression, []Token, error) {
	return p.parseBinary(tokens, TokenGreaterThan, p.parseGreaterEqual)
}

func (p *Parser) parseGreaterEqual(tokens []Token) (Expression, []Token, error) {
	return p.parseBinary(tokens, TokenGreaterEqual, p.parseLessThan)
}

func (p *Parser) parseLessThan(tokens []Token) (Expression, []Token, error) {
	return p.parseBinary(tokens, TokenLessThan, p.parseLessEqual)
}

func (p *Parser) parseLessEqual(tokens []Token) (Expression, []Token, error) {
	return p.parseBinary(tokens, TokenLessEqual, p.parseMinus)
}

func (p *Parser) parseMinus(tokens []Token) (Expression, []Token, error) {
	return p.parseBinary(tokens, TokenMinus, p.parsePlus)
}

func (p *Parser) parsePlus(tokens []Token) (Expression, []Token, error) {
	return p.parseBinary(tokens, TokenPlus, p.parseDivide)
}

func (p *Parser) parseDivide(tokens []Token) (Expression, []Token, error) {
	return p.parseBinary(tokens, TokenSlash, p.parseMultiply)
}

func (p *Parser) parseMultiply(tokens []Token) (Expression, []Token, error) {
	return p.parseBinary(tokens, TokenAsterisk, p.parseStringConcat)
}

func (p *Parser) parseStringConcat(tokens []Token) (Expression, []Token, error) {
	return p.parseBinary(tokens, TokenUnderscore, p.parseContainerAccess)
}

func (p *Parser) parseBinary(tokens []Token, tokenType TokenType, down func([]Token) (Expression, []Token, error)) (Expression, []Token, error) {
	left, next, err := down(tokens)
	if err != nil {
		return nil, nil, err
	}
	tokens = next

	for len(tokens) != 0 && tokens[0].Type == tokenType {
		right, next, err := down(tokens[1:])
		if err != nil {
			return nil, nil, err
		}

		left = &BinaryExpression{Left: left, Operator: tokens[0], Right: right}
		tokens = next
	}

	return left, tokens, nil
}

func (p *Parser) parsePrimary(tokens []Token) (Expression, []Token, error) {
	if len(tokens) == 0 {
		return nil, nil, fmt.Errorf("expected primary expression but reached end of data")
	}

	token := tokens[0]
	if token.Type == TokenString || token.Type == TokenNumber {
		return &LiteralExpression{Token: token}, tokens[1:], nil
	} else if token.Type == TokenIdentifier {
		if len(tokens) >= 2 && tokens[1].Type == TokenParenOpen {
			return p.parseFunctionCall(tokens)
		}

		// Variable access
		return &VariableExpression{Token: token}, tokens[1:], nil
	} else if token.Type == TokenParenOpen {
		expr, next, err := p.parseExpression(tokens[1:])
		if err != nil {
			return nil, nil, err
		}

		if len(next) == 0 || next[0].Type != TokenParenClose {
			tok := next[0]
			return nil, nil, fmt.Errorf("expect ')' after '(' and expression but got '%v' at %d:%d", tok.Type, tok.Line, tok.Col)
		}

		return expr, next[1:], nil
	}

	return nil, nil, fmt.Errorf("expected primary expression but got %s ('%s') at %d:%d", token.Type, token.Lexeme, token.Line, token.Col)
}

func (p *Parser) parseFunctionCall(tokens []Token) (Expression, []Token, error) {
	callToken := tokens[0]
	identifier := callToken.Lexeme

	builtin, found := builtins[identifier]
	if !found {
		tok := callToken
		return nil, nil, fmt.Errorf("no such builtin function '%s' at %d:%d", identifier, tok.Line, tok.Col)
	}

	tokens = tokens[2:] // Consume identifier and '('

	arguments := make([]Expression, 0)
	for len(tokens) > 0 {
		// TODO: remove duplication
		if tokens[0].Type == TokenParenClose {
			tokens = tokens[1:]
			break
		}

		expr, next, err := p.parseExpression(tokens)
		if err != nil {
			return nil, nil, err
		}

		arguments = append(arguments, expr)
		tokens = next
		if len(tokens) > 0 {
			if tokens[0].Type == TokenComma {
				tokens = tokens[1:]
			} else if tokens[0].Type != TokenParenClose {
				tok := tokens[0]
				return nil, nil, fmt.Errorf("expected ')' or ',' but got %s ('%s') at %d:%d", tok.Type, tok.Lexeme, tok.Line, tok.Col)
			}
		} else {
			return nil, nil, fmt.Errorf("expected ')' or ',' but got end of input")
		}
	}

	if len(arguments) != builtin.Arity && builtin.Arity != ArityVariadic {
		tok := tokens[0]
		return nil, nil, fmt.Errorf("expected %d arguments but got %d for function '%s' at %d:%d", builtin.Arity, len(arguments), identifier, tok.Line, tok.Col)
	}

	return &FunctionCallExpression{Token: callToken, FunctionName: callToken.Lexeme, Arguments: arguments}, tokens, nil
}
