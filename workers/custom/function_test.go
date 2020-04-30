package custom_test

import (
	"testing"

	"github.com/licaonfee/selina/workers"
	"github.com/licaonfee/selina/workers/custom"
)

func pass(in []byte) ([]byte, error) {
	return in, nil
}

func TestFunctionProcessCloseInput(t *testing.T) {
	f := custom.NewFunction(custom.FunctionOptions{Func: pass})
	if err := workers.ATProcessCloseInput(f); err != nil {
		t.Fatal(err)
	}
}

func TestFunctionProcessCloseOutput(t *testing.T) {
	f := custom.NewFunction(custom.FunctionOptions{Func: pass})
	if err := workers.ATProcessCloseOutput(f); err != nil {
		t.Fatal(err)
	}
}

func TestFunctionProcessCancel(t *testing.T) {
	f := custom.NewFunction(custom.FunctionOptions{Func: pass})
	if err := workers.ATProcessCancel(f); err != nil {
		t.Fatal(err)
	}
}
