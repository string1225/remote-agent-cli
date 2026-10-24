package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"
)

func readSetupToken(prompt, stdin bool) (string, error) {
	if prompt == stdin {
		return "", errors.New("choose --token-prompt or --token-stdin")
	}
	var secret []byte
	var err error
	if prompt {
		if !term.IsTerminal(int(os.Stdin.Fd())) {
			return "", errors.New("--token-prompt requires an interactive terminal; use --token-stdin for a pipe")
		}
		fmt.Fprint(os.Stderr, "粘贴 Token 后按 Enter（输入不回显）：")
		secret, err = term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Fprintln(os.Stderr)
	} else {
		secret, err = io.ReadAll(io.LimitReader(os.Stdin, 4097))
	}
	if err != nil {
		return "", errors.New("could not read token")
	}
	token := strings.TrimSpace(string(secret))
	if len(secret) > 4096 || token == "" || strings.ContainsAny(token, "\r\n") {
		return "", errors.New("invalid token input")
	}
	return token, nil
}
