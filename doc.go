// Copyright 2015, Joe Tsai. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE.md file.

package main

import "os"
import "fmt"
import "strings"
import "io/ioutil"
import "path"
import "path/filepath"
import "unicode"
import "unicode/utf8"
import "go/doc"
import "go/build"
import "go/token"
import "go/parser"

var args []string
var matchCase = false
var unexported = false
var showCmd = false

func handleDoc(t *tokenizer) {
	if tok, old := t.Next(); old {
		switch tok {
		case "-c", "--c":
			matchCase = true
		case "-u", "--u":
			unexported = true
		case "-cmd", "--cmd":
			showCmd = true
		default:
			args = append(args, tok)
		}
	} else {
		if strings.Index(tok, "-") == 0 {
			// go doc [-u] [-c] [-cmd]
			for _, s := range []string{"-u", "-c", "-cmd"} {
				if strings.Index(s, tok) == 0 {
					fmt.Println(s + " ")
				}
			}
		} else {
			switch len(args) {
			case 0:
				// go doc
				// go doc <sym>[.<method>]
				relPkg := false
				if cwd, err := os.Getwd(); err == nil {
					relPkg = suggestPackageContents(cwd, "", tok)
				}

				// go doc <pkg>
				if len(tok) > 0 || !relPkg {
					suggestPackages(tok)
				}

				// go doc [<pkg>].<sym>[.<method>]
				root, base := path.Split(tok)
				if i := strings.Index(base, "."); len(tok) > 0 && i > 0 {
					i += len(root)
					for _, pkgPath := range makePaths(tok[:i]) {
						tokRoot, tok := tok[:i+1], tok[i+1:]
						if suggestPackageContents(pkgPath, tokRoot, tok) {
							break
						}
					}
				}
			case 1:
				// go doc <pkg> <sym>[.<method>]
				for _, pkgPath := range makePaths(args[0]) {
					if suggestPackageContents(pkgPath, "", tok) {
						break
					}
				}
			}
		}
		return
	}
	handleDoc(t)
}

// Parse a package and suggest all symbols found in it.
func suggestPackageContents(pkgPath, tokRoot, tok string) bool {
	name, pkg := parsePackage(pkgPath)
	if pkg == nil || (name == "main" && !showCmd) {
		return false
	}

	for _, c := range pkg.Consts {
		for _, n := range c.Names {
			suggestSymbol(tokRoot, tok, n)
		}
	}
	for _, v := range pkg.Vars {
		for _, n := range v.Names {
			suggestSymbol(tokRoot, tok, n)
		}
	}
	for _, f := range pkg.Funcs {
		suggestSymbol(tokRoot, tok, f.Name)
	}
	for _, t := range pkg.Types {
		suggestSymbol(tokRoot, tok, t.Name)
		for _, m := range t.Methods {
			suggestSymbol(tokRoot, tok, t.Name+"."+m.Name)
		}
	}
	return true
}

// Print the symbol as a suggestion.
func suggestSymbol(tokRoot, tok, name string) {
	// Check if name is exported
	if !unexported {
		i := strings.LastIndex(name, ".")
		rune, _ := utf8.DecodeRuneInString(name[i+1:])
		if !unicode.IsUpper(rune) {
			return
		}
	}

	if strings.Index(name, tok) == 0 {
		fmt.Println(tokRoot + name + " ")
	} else if !matchCase {
		// Copied from cmd/doc/pkg.go.
		simpleFold := func(r rune) rune {
			for {
				r1 := unicode.SimpleFold(r)
				if r1 <= r {
					return r1
				}
				r = r1
			}
		}

		// Copied from cmd/doc/pkg.go.
		for _, u := range tok {
			p, w := utf8.DecodeRuneInString(name)
			name = name[w:]
			if u == p {
				continue
			}
			if unicode.IsLower(u) && simpleFold(u) == simpleFold(p) {
				continue
			}
			return
		}
		fmt.Println(tokRoot + tok + name + " ")
	}
}

// Suggest packages that tok could possibly be pointing toward.
func suggestPackages(tok string) {
	root, base := path.Split(tok)
	if isDir(tok) {
		root, base = tok, ""
	}

	for _, dirPath := range makePaths(root) {
		for _, dir := range getDirs(dirPath) {
			if strings.Index(dir, base) != 0 {
				// Skip paths don't match the partial token
				continue
			}
			if strings.Index(dir, ".") == 0 && len(base) == 0 {
				// Skip hidden folders unless there is a partial match
				continue
			}
			path := path.Join(dirPath, dir)
			if isPackage(path) {
				fmt.Println(root + dir + " ")
			}
			if hasPackages(path) {
				fmt.Println(root + dir + string(os.PathSeparator))
			}
		}
	}
}

// Expand GOROOT and GOPATH with respect to some dirPath.
func makePaths(dirPath string) (paths []string) {
	// TODO(jtsai): Can dirPath be an absolute path?
	for _, dir := range filepath.SplitList(build.Default.GOPATH) {
		if dir != "" {
			paths = append(paths, path.Join(dir, "src", dirPath))
		}
	}
	if dir := build.Default.GOROOT; dir != "" {
		paths = append(paths, path.Join(dir, "src", dirPath))
	}
	return
}

// Check if dirPath points to a directory.
func isDir(dirPath string) bool {
	return len(dirPath) > 0 && dirPath[len(dirPath)-1] == os.PathSeparator
}

// Get all directories within the directory at dirPath.
func getDirs(dirPath string) (dirs []string) {
	fis, _ := ioutil.ReadDir(dirPath)
	for _, fi := range fis {
		if fi.Mode()&os.ModeSymlink > 0 {
			_fi, err := os.Stat(path.Join(dirPath, fi.Name()))
			if err == nil {
				fi = _fi
			}
		}
		if fi.IsDir() {
			dirs = append(dirs, fi.Name())
		}
	}
	return
}

// Check if the path is a Go package. This function simply checks if at least
// one Go file exists in the directory and could be wrong.
func isPackage(path string) bool {
	fis, _ := ioutil.ReadDir(path)
	for _, fi := range fis {
		if strings.HasSuffix(fi.Name(), ".go") {
			return true
		}
	}
	return false
}

// Check if this directory contains other Go packages. This function simply
// checks if other directories exist within this directory and could be wrong.
func hasPackages(path string) bool {
	return len(getDirs(path)) > 0
}

// Parse a package at a given dir.
// This logic is a stripped down version of cmd/doc/pkg.go.
func parsePackage(dir string) (_ string, _ *doc.Package) {
	if dir == "" {
		return
	}

	pkg, err := build.ImportDir(dir, build.ImportComment)
	if err != nil {
		return
	}

	fs := token.NewFileSet()
	include := func(info os.FileInfo) bool {
		for _, name := range pkg.GoFiles {
			if name == info.Name() {
				return true
			}
		}
		for _, name := range pkg.CgoFiles {
			if name == info.Name() {
				return true
			}
		}
		return false
	}
	pkgs, err := parser.ParseDir(fs, pkg.Dir, include, parser.ParseComments)
	if err != nil {
		return
	}
	if len(pkgs) != 1 {
		return
	}
	astPkg := pkgs[pkg.Name]

	docPkg := doc.New(astPkg, pkg.ImportPath, doc.AllDecls)
	for _, typ := range docPkg.Types {
		docPkg.Consts = append(docPkg.Consts, typ.Consts...)
		docPkg.Vars = append(docPkg.Vars, typ.Vars...)
		docPkg.Funcs = append(docPkg.Funcs, typ.Funcs...)
	}

	return pkg.Name, docPkg
}
