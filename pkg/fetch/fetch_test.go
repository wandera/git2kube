package fetch

import (
	"go/types"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	"testing"
)

func TestNewAuth(t *testing.T) {
	cases := []struct {
		name   string
		git    string
		result transport.AuthMethod
	}{
		{
			name:   "No Auth",
			git:    "https://github.com/WanderaOrg/git2kube.git",
			result: nil,
		},
		{
			name: "Basic Auth",
			git:  "https://test:testpass@github.com/WanderaOrg/git2kube.git",
			result: &http.BasicAuth{
				Username: "test",
				Password: "testpass",
			},
		},
	}

	for _, c := range cases {
		m, err := NewAuth(c.git)
		if err != nil {
			t.Errorf("%s case failed: failed to parse git uri", c.name)
		}

		switch m.(type) {
		case *types.Nil:
			if c.result != nil {
				t.Errorf("%s case failed: result should have been nil but got %s instead", c.name, c.result)
			}
		case *http.BasicAuth:
			g := m.(*http.BasicAuth)
			w := c.result.(*http.BasicAuth)
			if g.Username != w.Username || g.Password != w.Password {
				t.Errorf("%s case failed: result mismatch expected %s but got %s instead", c.name, c.result, m)
			}
		}

	}
}
