package txl_test

import (
	"testing"

	"github.com/matsuyoshi30/txl/passes/txl"
	"golang.org/x/tools/go/analysis/analysistest"
)

func Test(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, txl.Analyzer, "a")
}
