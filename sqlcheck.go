/*
MIT License

Copyright (c) 2018 Luka Napotnik

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package sqlcheck

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type token int

// syntax rules define what must be before a token (ie. the lvalue)
var syntaxRules = map[token][]token{
	tokenSelect:    []token{tokenEOF, tokenLeftBracket},
	tokenIdent:     []token{tokenSelect, tokenFrom, tokenWhere, tokenLessThan, tokenMoreThan, tokenComma},
	tokenFrom:      []token{tokenIdent, tokenAsterisk},
	tokenWhere:     []token{tokenIdent},
	tokenLessThan:  []token{tokenIdent, tokenNumeral},
	tokenMoreThan:  []token{tokenIdent, tokenNumeral},
	tokenEqualTo:   []token{tokenIdent, tokenNumeral},
	tokenAsterisk:  []token{tokenSelect},
	tokenNumeral:   []token{tokenWhere, tokenLessThan, tokenMoreThan, tokenEqualTo},
	tokenSemicolon: []token{tokenIdent, tokenNumeral},
	tokenComma:     []token{tokenIdent},
}

type astItem struct {
	tok      token
	pos      int
	value    string
	children []*astItem
}

type astRoot struct {
	children []*astItem
}

func (root *astRoot) append(newItem *astItem) {
	root.children = append(root.children, newItem)
}

func (root *astRoot) print() {
	for i := range root.children {
		fmt.Printf("%v (%v)\n", root.children[i].tok, root.children[i].value)
	}
}

const (
	tokenIllegal token = iota
	tokenEOF
	tokenLeftBracket
	tokenRightBracket
	tokenAsterisk
	tokenComma
	tokenSemicolon
	tokenSelect
	tokenFrom
	tokenWhere
	tokenLessThan
	tokenMoreThan
	tokenEqualTo
	tokenNumeral
	tokenIdent
)

func (t token) String() string {
	switch t {
	case tokenEOF:
		return "EOF"
	case tokenLeftBracket:
		return "("
	case tokenRightBracket:
		return ")"
	case tokenAsterisk:
		return "*"
	case tokenComma:
		return ","
	case tokenSelect:
		return "SELECT"
	case tokenFrom:
		return "FROM"
	case tokenWhere:
		return "WHERE"
	case tokenLessThan:
		return "<"
	case tokenMoreThan:
		return ">"
	case tokenEqualTo:
		return "="
	case tokenNumeral:
		return "(0..9)"
	case tokenIdent:
		return "identifier"
	}
	return "illegal"
}

type sqlLexer struct {
	scanner  *bufio.Scanner
	curToken string
	ast      *astRoot
}

func isWhitespace(s string) bool {
	if s == " " || s == "\t" || s == "\n" || s == "\r" {
		return true
	}
	return false
}

func isNumeral(s string) bool {
	if s[0] > 47 && s[0] < 58 {
		return true
	}
	return false
}

func isIdent(s string) bool {
	if (s[0] > 64 && s[0] < 91) || (s[0] > 96 && s[0] < 123) {
		return true
	}
	return false
}

func (l *sqlLexer) check(tok token) error {
	if len(l.ast.children) == 0 {
		return nil
	}
	var (
		lValues   = syntaxRules[tok]
		prevToken = l.ast.children[len(l.ast.children)-1].tok
	)

	validLvalue := false
	for i := range lValues {
		//fmt.Printf("token %v, lValues: %v\n", tok, lValues)
		if prevToken == lValues[i] {
			validLvalue = true
			break
		}
	}
	if !validLvalue {
		return fmt.Errorf("left side of %q must be one of %v", tok, lValues)
	}
	return nil
}

func (l *sqlLexer) getAstItem(v string) *astItem {
	vLower := strings.ToLower(v)
	t := tokenIllegal
	switch vLower {
	case "select":
		t = tokenSelect
	case "from":
		t = tokenFrom
	case "where":
		t = tokenWhere
	case "*":
		t = tokenAsterisk
	case ",":
		t = tokenComma
	case ";":
		t = tokenSemicolon
	case "<":
		t = tokenLessThan
	case ">":
		t = tokenMoreThan
	case "=":
		t = tokenEqualTo
	default:
		if _, err := strconv.Atoi(v); err == nil {
			t = tokenNumeral
		} else {
			t = tokenIdent
		}
	}
	return &astItem{
		tok:   t,
		value: v,
	}
}

func (l *sqlLexer) Verify() error {
	l.scanner.Split(bufio.ScanRunes)
	for l.scanner.Scan() {
		text := l.scanner.Text()
		if isWhitespace(text) {
			if len(l.curToken) > 0 {
				// We have a token
				item := l.getAstItem(l.curToken)
				//fmt.Printf("item %v\n", item.value)
				if err := l.check(item.tok); err != nil {
					return err
				}
				l.ast.append(item)
				l.curToken = ""
			}
			continue
		}
		if isNumeral(text) && len(l.curToken) == 0 {
			l.curToken += text
			continue
		}
		// identifiers can't start with numerals but can contain them
		if (isNumeral(text) && len(l.curToken) > 0) || isIdent(text) {
			l.curToken += text
			continue
		}
		if len(l.curToken) > 0 {
			// We have a token
			item := l.getAstItem(l.curToken)
			//fmt.Printf("item %v\n", item.value)
			if err := l.check(item.tok); err != nil {
				return err
			}
			l.ast.append(item)
			l.curToken = ""
		}
		item := l.getAstItem(text)
		//fmt.Printf("item %v\n", item.value)
		if err := l.check(item.tok); err != nil {
			return err
		}
		l.ast.append(item)
	}
	return nil
}

func NewCheckBuffered(buf io.Reader) *sqlLexer {
	return &sqlLexer{
		ast:     &astRoot{},
		scanner: bufio.NewScanner(buf),
	}
}

func NewCheck(stmt string) *sqlLexer {
	return &sqlLexer{
		ast:     &astRoot{},
		scanner: bufio.NewScanner(bytes.NewBufferString(stmt)),
	}
}
