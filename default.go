// Copyright 2015, Joe Tsai. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE.md file.

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

func handleDefault(t *tokenizer) {
	if tok, old := t.Next(); old {
		handleDefault(t)
	} else {
		// go <cmd> <pkg>
		suggestPackages(tok)

		// go <cmd> <path>
		suggestPaths(tok)
	}
}

func suggestPaths(tok string) (cnt int) {
	root, base := path.Split(tok)
	if isDir(tok) {
		root, base = tok, ""
	}

	fis, _ := ioutil.ReadDir(root)
	if root == "" {
		fis, _ = ioutil.ReadDir(".")
	}

	for _, fi := range fis {
		if strings.Index(fi.Name(), base) != 0 {
			// Skip paths don't match the partial token
			continue
		}
		if strings.Index(fi.Name(), ".") == 0 && len(base) == 0 {
			// Skip hidden folders unless there is a partial match
			continue
		}
		if fi.IsDir() {
			fmt.Println(root + fi.Name() + string(os.PathSeparator))
			cnt++
		} else {
			fmt.Println(root + fi.Name() + " ")
			cnt++
		}
	}
	return cnt
}
