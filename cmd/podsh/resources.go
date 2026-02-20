package main

import (
	"fmt"
	"os"
)

const (
	podsh = "podsh"
	short = "podsh is the ssh gateway for pods on kthcloud"
	long  = `                 __    __ 
   ___  ___  ___/ /__ / / 
  / _ \/ _ \/ _  (_-</ _ \
 / .__/\___/\_,_/___/_//_/
/_/                      `
)

var version = ""

func banner() {
	fmt.Fprintf(os.Stderr, "%s\nVersion: %s\n", long, version)
}
