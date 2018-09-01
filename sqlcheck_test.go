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
	"testing"
)

func TestSqlLex(t *testing.T) {
	cases := []struct {
		name      string
		statement string
	}{
		{"SELECT multiple values", "SELECT abc, def FROM names;"},
		{"SELECT multiple values WHERE", "SELECT name FROM names WHERE age > 20;"},
	}
	for _, tCase := range cases {
		t.Run(tCase.name, func(t *testing.T) {
			l := NewCheck(tCase.statement)
			if err := l.Verify(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestSqlLexFail(t *testing.T) {
	cases := []struct {
		name      string
		statement string
	}{
		{"missing statement for SELECT", "SELECT FROM names;"},
		{"numeral for select statement", "SELECT 123 FROM names;"},
		{"SELECT multiple values WHERE invalid", "SELECT name FROM names WHERE > 20;"},
	}
	for _, tCase := range cases {
		t.Run(tCase.name, func(t *testing.T) {
			l := NewCheck(tCase.statement)
			if err := l.Verify(); err == nil {
				t.Fatal("test should fail")
			}
		})
	}

}

func BenchmarkSqlLex(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if err := NewCheck("SELECT name FROM names WHERE age > 20;").Verify(); err != nil {
			b.Fatal(err)
		}
	}
}
