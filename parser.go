package main

import (
	"fmt"
)

type Env map[string]any
type Statement func(Env) error
type Expression func(Env) (any, error)

func parse(tokens []Token) (statements []Statement, err error) {
	for len(tokens) > 0 {
		stmt, next, err := parseStatement(tokens)
		if err != nil {
			return nil, err
		}
		if stmt != nil {
			statements = append(statements, stmt)
		}
		tokens = next
	}

	return
}

func parseStatement(tokens []Token) (stmt Statement, next []Token, err error) {
	if tokens[0].Type == TokenNewline {
		// Skip over empty lines
		return nil, tokens[1:], nil
	}

	if tokens[0].Type == TokenIf {
		stmt, next, err = parseIfStatement(tokens)
		if err != nil {
			return nil, next, err
		}
	} else if tokens[0].Type == TokenWhile {
		stmt, next, err = parseWhileStatement(tokens)
		if err != nil {
			return nil, next, err
		}
	} else if len(tokens) >= 2 && tokens[0].Type == TokenIdentifier && tokens[1].Type == TokenEquals {
		stmt, next, err = parseAssignmentStatement(tokens)
		if err != nil {
			return nil, next, err
		}
	} else {
		expr, nextTokens, err := parseExpression(tokens)
		if err != nil {
			return nil, nil, err
		}
		stmt = func(env Env) error {
			_, err := expr(env) /* Discard return value */
			return err
		}

		next = nextTokens
	}

	if len(next) != 0 && next[0].Type != TokenNewline && next[0].Type != TokenBraceClose {
		return nil, nil, fmt.Errorf("expected newline after statement but got %s ('%s')", next[0].Type, next[0].Lexeme)
	}

	if len(next) != 0 && next[0].Type == TokenNewline {
		// Consume newline
		next = next[1:]
	}
	return stmt, next, nil
}

func parseBlock(tokens []Token, typ string) (Statement, []Token, error) {
	next := tokens

	// TODO: parse blocks
	if len(next) == 0 || next[0].Type != TokenBraceOpen {
		if len(next) != 0 {
			next = next[1:]
		}
		return nil, nil, fmt.Errorf("expected '{' after %s expression", typ)
	}

	next = next[1:]
	statements := make([]Statement, 0)
	for len(next) != 0 && next[0].Type != TokenBraceClose {
		stmt, next2, err := parseStatement(next)
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
		return nil, nil, fmt.Errorf("expected '}' after %s statements", typ)
	}

	return func(env Env) error {
		for _, stmt := range statements {
			if err := stmt(env); err != nil {
				return err
			}
		}
		return nil
	}, next[1:], nil
}

func parseIfStatement(tokens []Token) (Statement, []Token, error) {
	expr, next, err := parseExpression(tokens[1:])
	if err != nil {
		return nil, nil, err
	}

	block, next, err := parseBlock(next, "if")
	if err != nil {
		return nil, nil, err
	}

	return func(env Env) error {
		v, err := expr(env)
		if err != nil {
			return err
		}
		if isWeirdlyTrue(v) {
			return block(env)
		}
		return nil
	}, next, nil
}

func parseWhileStatement(tokens []Token) (Statement, []Token, error) {
	expr, next, err := parseExpression(tokens[1:])
	if err != nil {
		return nil, nil, err
	}

	block, next, err := parseBlock(next, "while")
	if err != nil {
		return nil, nil, err
	}

	return func(env Env) error {
		for {
			v, err := expr(env)
			if err != nil {
				return err
			}
			if !isWeirdlyTrue(v) {
				break
			}

			if err := block(env); err != nil {
				return err
			}
		}
		return nil
	}, next, nil
}

func parseAssignmentStatement(tokens []Token) (Statement, []Token, error) {
	if len(tokens) < 3 {
		return nil, nil, fmt.Errorf("expected '=' and expression after identifier")
	} else if tokens[1].Type != TokenEquals {
		return nil, nil, fmt.Errorf("expected '=' after identifier")
	}

	expr, next, err := parseExpression(tokens[2:])
	if err != nil {
		return nil, nil, err
	}

	identifier := tokens[0].Lexeme
	return func(env Env) error {
		v, err := expr(env)
		if err != nil {
			return err
		}
		env[identifier] = v
		return nil
	}, next, nil
}

func parseExpression(tokens []Token) (Expression, []Token, error) {
	return parseEqualEqual(tokens)
}

func parseEqualEqual(tokens []Token) (Expression, []Token, error) {
	return parseBinary(tokens, TokenEqualEqual, parseNotEqual, func(l, r int) int { return boolToInt(l == r) })
}

func parseNotEqual(tokens []Token) (Expression, []Token, error) {
	return parseBinary(tokens, TokenNotEqual, parseGreaterThan, func(l, r int) int { return boolToInt(l != r) })
}

func parseGreaterThan(tokens []Token) (Expression, []Token, error) {
	return parseBinary(tokens, TokenGreaterThan, parseGreaterEqual, func(l, r int) int { return boolToInt(l > r) })
}

func parseGreaterEqual(tokens []Token) (Expression, []Token, error) {
	return parseBinary(tokens, TokenGreaterEqual, parseLessThan, func(l, r int) int { return boolToInt(l >= r) })
}

func parseLessThan(tokens []Token) (Expression, []Token, error) {
	return parseBinary(tokens, TokenLessThan, parseLessEqual, func(l, r int) int { return boolToInt(l < r) })
}

func parseLessEqual(tokens []Token) (Expression, []Token, error) {
	return parseBinary(tokens, TokenLessEqual, parseMinus, func(l, r int) int { return boolToInt(l <= r) })
}

func parseMinus(tokens []Token) (Expression, []Token, error) {
	return parseBinary(tokens, TokenMinus, parsePlus, func(l, r int) int { return l - r })
}

func parsePlus(tokens []Token) (Expression, []Token, error) {
	return parseBinary(tokens, TokenPlus, parseDivide, func(l, r int) int { return l + r })
}

func parseDivide(tokens []Token) (Expression, []Token, error) {
	return parseBinary(tokens, TokenSlash, parseMultiply, func(l, r int) int { return l / r })
}

func parseMultiply(tokens []Token) (Expression, []Token, error) {
	return parseBinary(tokens, TokenAsterisk, parseUnderscore, func(l, r int) int { return l * r })
}

func parseUnderscore(tokens []Token) (Expression, []Token, error) {
	left, next, err := parsePrimary(tokens)
	if err != nil {
		return nil, nil, err
	}
	tokens = next

	for len(tokens) != 0 && tokens[0].Type == TokenUnderscore {
		right, next, err := parsePrimary(tokens[1:])
		if err != nil {
			return nil, nil, err
		}

		leftHand := left
		left = func(env Env) (any, error) {
			var lv, rv any
			var ls, rs string
			var err error
			var ok bool

			if lv, err = leftHand(env); err != nil {
				return nil, err
			}

			if ls, ok = lv.(string); !ok {
				return nil, fmt.Errorf("left-hand operand of '_' should be a string but was '%v'", lv)
			}

			if rv, err = right(env); err != nil {
				return nil, err
			}

			if rs, ok = rv.(string); !ok {
				return nil, fmt.Errorf("right-hand operand of '_' should be a string but was '%v'", rv)
			}

			return ls + rs, nil
		}

		tokens = next
	}

	return left, tokens, nil
}

func parseBinary(tokens []Token, tokenType TokenType, down func([]Token) (Expression, []Token, error), op func(int, int) int) (Expression, []Token, error) {
	left, next, err := down(tokens)
	if err != nil {
		return nil, nil, err
	}
	tokens = next

	for len(tokens) != 0 && tokens[0].Type == tokenType {
		lexeme := tokens[0].Lexeme

		right, next, err := down(tokens[1:])
		if err != nil {
			return nil, nil, err
		}

		leftHand := left
		left = func(env Env) (any, error) {
			var lv, rv any
			var li, ri int
			var err error
			var ok bool

			if lv, err = leftHand(env); err != nil {
				return nil, err
			}

			if li, ok = lv.(int); !ok {
				return nil, fmt.Errorf("left-hand operand of '%s' should be an int but was '%v'", lexeme, lv)
			}

			if rv, err = right(env); err != nil {
				return nil, err
			}

			if ri, ok = rv.(int); !ok {
				return nil, fmt.Errorf("right-hand operand of '%s' should be an int but was '%v'", lexeme, rv)
			}

			return op(li, ri), nil
		}

		tokens = next
	}

	return left, tokens, nil
}

func parsePrimary(tokens []Token) (Expression, []Token, error) {
	if len(tokens) == 0 {
		return nil, nil, fmt.Errorf("expected primary expression but reached end of data")
	}

	token := tokens[0]
	if token.Type == TokenString {
		value := tokens[0].Literal
		return func(Env) (any, error) { return value, nil }, tokens[1:], nil
	} else if token.Type == TokenNumber {
		value := tokens[0].Literal
		return func(Env) (any, error) { return value, nil }, tokens[1:], nil
	} else if token.Type == TokenIdentifier {
		identifier := token.Lexeme

		if len(tokens) >= 2 && tokens[1].Type == TokenParenOpen {
			return parseFunctionCall(tokens)
		}

		// Variable access
		return func(env Env) (any, error) {
			val, found := env[identifier]
			if found {
				return val, nil
			}
			return nil, fmt.Errorf("undefined variable '%s'", identifier)
		}, tokens[1:], nil
	}

	return nil, nil, fmt.Errorf("expected primary expression but got %s ('%s')", token.Type, token.Lexeme)
}

func parseFunctionCall(tokens []Token) (Expression, []Token, error) {
	identifier := tokens[0].Lexeme

	builtin, found := builtins[identifier]
	if !found {
		return nil, nil, fmt.Errorf("no such builtin function '%s'", identifier)
	}

	tokens = tokens[2:] // Consume '('

	arguments := make([]Expression, 0)
	for len(tokens) > 0 {
		// TODO: remove duplication
		if tokens[0].Type == TokenParenClose {
			tokens = tokens[1:]
			break
		}

		expr, next, err := parseExpression(tokens)
		if err != nil {
			return nil, nil, err
		}

		arguments = append(arguments, expr)
		tokens = next
		if len(tokens) > 0 {
			if tokens[0].Type == TokenComma {
				tokens = tokens[1:]
			} else if tokens[0].Type != TokenParenClose {
				return nil, nil, fmt.Errorf("expected ')' or ',' but got %s ('%s')", tokens[0].Type, tokens[0].Lexeme)
			}
		} else {
			return nil, nil, fmt.Errorf("expected ')' or ',' but got end of input")
		}
	}

	if len(arguments) != builtin.Arity && builtin.Arity != ArityVariadic {
		return nil, nil, fmt.Errorf("expected %d arguments but got %d for function '%s'", builtin.Arity, len(arguments), identifier)
	}

	return func(env Env) (any, error) { return builtin.Func(env, arguments) }, tokens, nil
}

func isWeirdlyTrue(v any) bool {
	return v != 0
}

func boolToInt(b bool) int {
	if b {
		return 1
	} else {
		return 0
	}
}

func bla(left Expression, right Expression, env Env, tok string, op func(int, int) int) (any, error) {
	var lv, rv any
	var li, ri int
	var err error
	var ok bool

	if lv, err = left(env); err != nil {
		return nil, err
	}

	if li, ok = lv.(int); !ok {
		return nil, fmt.Errorf("left-hand operand of '%s' should be a string but was '%v'", tok, lv)
	}

	if rv, err = right(env); err != nil {
		return nil, err
	}

	if ri, ok = rv.(int); !ok {
		return nil, fmt.Errorf("right-hand operand of '%s' should be a string but was '%v'", tok, rv)
	}

	return op(li, ri), nil
}
