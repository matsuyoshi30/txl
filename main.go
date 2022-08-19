package main

import (
	"github.com/matsuyoshi30/txl/passes/txl"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(txl.Analyzer)
}
