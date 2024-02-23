package geomys

import "testing"

var cases = map[string]string{
	"connectrpc.com/connect":     "com_connectrpc_connect",
	"github.com/golang/protobuf": "com_github_golang_protobuf",
	"github.com/google/go-cmp":   "com_github_google_go_cmp",
	"golang.org/x/mod":           "org_golang_x_mod",
}

func TestCanonicalize(t *testing.T) {
	for test, expected := range cases {
		if output := CanonicalizeModuleName(test); output != expected {
			t.Errorf("Output %q not equal to expected %q", output, expected)
		}
	}
}
