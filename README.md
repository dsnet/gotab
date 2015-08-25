# Tab completion for Go binary #

## Introduction ##

This program provides tab completion information for the Go binary. Currently,
this program only works with the *doc* tool in the *go* binary for version
1.5.

Furthermore, this only works for bash, but thin wrappers can be used to make it
work with other shells.

## Installation ##

1. Build the gotab binary and place it in your *PATH*.
2. Add the following to your bashrc file:
```complete -C gotab -o nospace go```

## Usage ##

Use the go binary and hit tab to auto-complete if possible.

Thus, when you type the following and hit tab:
```bash
$ go doc runtime CP
```

It will auto-complete to the following:
```bash
$ go doc runtime CPUProfile
```

If there are more than one possible completion, then they will be listed:
```bash
$ go doc runtime Read
ReadMemStats   ReadTrace
```
