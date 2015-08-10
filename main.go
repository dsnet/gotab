// Copyright 2015, Joe Tsai. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE.md file.

// GoTab is helper program that provides information for tab completion on the
// Go command.
//
// To install:
//	1. Build the gotab binary and place it in your PATH.
//	2. Add the following to your bashrc file:
//		complete -C gotab -o nospace go
//
// This program works by taking advantage of the built-in complete command for
// bash. The complete command above informs bash to call gotab for suggestions
// anytime tab completion is used on the go binary. When gotab is called by
// bash, it will be provided with the environment variables COMP_LINE and
// COMP_POINT where COMP_LINE contains the user's current command line and
// COMP_POINT points to the location of the user's cursor. This program ignores
// any characters in COMP_LINE that lies beyond COMP_POINT.
package main

import "os"
import "fmt"
import "path"
import "strings"
import "strconv"

const cmd = "go"

var tools = []string{
	"build", "clean", "doc", "env", "fix", "fmt", "generate", "get", "install",
	"list", "run", "test", "tool", "version", "vet", "help",
}

func main() {
	if t := newTokenizer(); t != nil {
		if tok, old := t.Next(); old {
			switch tok {
			case "doc":
				handleDoc(t)
			default:
				handleDefault(t)
			}
		} else {
			// Suggest a tool
			for _, tool := range tools {
				if strings.Index(tool, tok) == 0 {
					fmt.Println(tool + " ")
				}
			}
		}
		os.Exit(0)
	}

	fmt.Fprintf(os.Stderr, "Error: COMP_LINE and COMP_POINT environment variables not set.\n")
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "Do not call %s directly. Instead, do the following:\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "1. Place this binary in your PATH.\n")
	fmt.Fprintf(os.Stderr, "2. Place this line in your bashrc file:\n")
	fmt.Fprintf(os.Stderr, "	complete -C %s -o nospace go\n", path.Base(os.Args[0]))

	os.Exit(1)
}

type tokenizer struct {
	line  string
	point int
	end   bool
}

func newTokenizer() *tokenizer {
	line := os.Getenv("COMP_LINE")
	point, err := strconv.Atoi(os.Getenv("COMP_POINT"))
	if err != nil || point > len(line) || strings.Index(line, cmd+" ") != 0 {
		return nil
	}
	t := &tokenizer{line, point, false}
	t.Next() // Consume first token
	return t
}

// Iterate through each token on the command-line until we reach a token where
// the pointer is at. Once that happens, old will be set to true.
func (t *tokenizer) Next() (tok string, old bool) {
	var i int
	// TODO(jtsai): Handle shell escaping.

	// Strip any white spaces
	for i = 0; i < len(t.line) && t.line[i] == ' '; i++ {
	}
	t.line, t.point = t.line[i:], t.point-i

	// Find the next token
	for i = 0; i < len(t.line) && t.line[i] != ' '; i++ {
	}
	pt := t.point
	tok, t.line, t.point = t.line[:i], t.line[i:], t.point-i

	// Truncate token if pointer is inside token
	old = bool(pt > len(tok))
	if pt <= len(tok) {
		if pt < 0 {
			pt = 0
		}
		tok = tok[:pt]
		t.line, t.point = "", 0
	}
	return
}
