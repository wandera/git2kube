package fetch

import (
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	gitssh "gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
	"testing"
)

func TestNewAuth(t *testing.T) {
	cases := []struct {
		name   string
		git    string
		key    string
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
		{
			name: "SSH wit private key",
			git:  "git@github.com:WanderaOrg/git2kube.git",
			key:  "testdata/dummy.key",
			result: &gitssh.PublicKeys{
				User: "git",
				Signer: nil,
			},
		},
		{
			name: "HTTP url with private key",
			git:  "https://github.com/WanderaOrg/git2kube.git",
			key:  "/tmp/i_am_not_here.key",
			result: nil,
		},
		{
			name: "SSH url without private key",
			git:  "git@github.com:WanderaOrg/git2kube.git",
			result: nil,
		},								
	}

	for _, c := range cases {
		m, err := NewAuth(c.git, c.key)
		if err != nil {
			t.Errorf("%s case failed: %s", c.name, err)
		}
		
		if m == nil && c.result != nil {
			t.Errorf("%s case failed: result should have been %s but got nil instead", c.name, c.result)
		} else if m != nil && c.result == nil {
			t.Errorf("%s case failed: result should have been nil but got %s instead", c.name, c.result)
		}
		
		switch m.(type) {
		case *http.BasicAuth:
			g := m.(*http.BasicAuth)
			w := c.result.(*http.BasicAuth)
			if g.Username != w.Username || g.Password != w.Password {
				t.Errorf("%s case failed: result mismatch expected %s but got %s instead", c.name, c.result, m)
			}
		case *gitssh.PublicKeys:
			g := m.(*gitssh.PublicKeys)
			w := c.result.(*gitssh.PublicKeys)
			if g.User != w.User {
				t.Errorf("%s case failed: result mismatch expected %s but got %s instead", c.name, c.result, m)
			} else if g.Signer == nil {
				t.Errorf("%s case failed: result mismatch expected Signer but got nil instead", c.name)
			}
		}
	}
}
