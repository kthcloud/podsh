package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path"
	"strings"
)

var heuristicCommonShellLookup = []string{
	// Bash
	"/usr/bin/bash",
	"/bin/bash",
	"/usr/local/bin/bash",

	// Sh (POSIX)
	"/usr/bin/sh",
	"/bin/sh",

	// Zsh (default on macOS)
	"/usr/bin/zsh",
	"/bin/zsh",
	"/usr/local/bin/zsh",

	// Dash (Debian/Ubuntu /bin/sh alternative)
	"/usr/bin/dash",
	"/bin/dash",

	// Ash (BusyBox / Alpine)
	"/bin/ash",
	"/usr/bin/ash",

	// BusyBox
	"/bin/busybox",
	"/usr/bin/busybox",

	// Fish
	"/usr/bin/fish",
	"/usr/local/bin/fish",

	// Ksh
	"/usr/bin/ksh",
	"/bin/ksh",

	// Csh / Tcsh
	"/usr/bin/csh",
	"/bin/csh",
	"/usr/bin/tcsh",
	"/bin/tcsh",
}

// USAGE getsh [root] [userid] [pid]
func main() {
	root := ""
	uid := "0"
	pid := "1"
	for i, arg := range os.Args[1:] {
		switch i {
		case 0:
			root = strings.TrimSuffix(root, "/")
		case 1:
			uid = arg
		case 2:
			pid = arg
		}
	}

	if shell, err := shellFromProcEnv(pid); err == nil && shell == "" && isExecutable(path.Join(root, shell)) {
		os.Stdout.Write([]byte(shell))
		return
	}

	shell := shellFromPasswd(root, uid)
	if shell != "" && isExecutable(path.Join(root, shell)) {
		os.Stdout.Write([]byte(shell))
		return
	}

	for _, sh := range heuristicCommonShellLookup {
		shell := path.Join(root, sh)
		if isExecutable(shell) {
			os.Stdout.Write([]byte(sh))
			return
		}
	}
	os.Exit(1)
}

func shellFromProcEnv(pid string) (string, error) {
	data, err := os.ReadFile("/proc/" + pid + "/environ")
	if err != nil {
		return "", err
	}

	// environ is null-byte separated
	envVars := bytes.SplitSeq(data, []byte{0})

	for e := range envVars {
		if bytes.HasPrefix(e, []byte("SHELL=")) {
			return string(e[len("SHELL="):]), nil
		}
	}

	return "", fmt.Errorf("SHELL not found")
}

func isExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir() && info.Mode()&0o111 != 0
}

func shellFromPasswd(root string, userID string) string {
	file, err := os.Open(root + "/etc/passwd")
	if err != nil {
		return ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Split(line, ":")
		if len(fields) < 7 {
			continue
		}
		uid := fields[2]
		if uid == userID {
			return fields[6]
		}
	}
	return ""
}
