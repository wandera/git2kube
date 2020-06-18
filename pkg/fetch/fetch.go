package fetch

import (
	"errors"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	gitssh "github.com/go-git/go-git/v5/plumbing/transport/ssh"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"os"
	"strings"
)

// Fetcher fetching data from remote
type Fetcher interface {
	Fetch() (*object.Commit, error)
}

type fetcher struct {
	url       string
	directory string
	branch    string
	auth      transport.AuthMethod
}

// NewFetcher creates new Fetcher
func NewFetcher(url string, directory string, branch string, auth transport.AuthMethod) Fetcher {
	fetcher := &fetcher{
		url:       url,
		directory: directory,
		branch:    branch,
		auth:      auth,
	}
	return fetcher
}

// Fetch from remote
func (f *fetcher) Fetch() (*object.Commit, error) {
	err := os.RemoveAll(f.directory)
	if err != nil {
		log.Errorf("Failed to clean the folder '%s': %v", f.branch, err)
		return nil, err
	}

	log.Infof("Cloning repository in folder '%s' ...", f.directory)
	r, err := git.PlainClone(f.directory, false, &git.CloneOptions{
		URL:           f.url,
		Auth:          f.auth,
		Depth:         1,
		ReferenceName: plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", f.branch)),
	})
	if err != nil {
		log.Errorf("Failed to clone '%s': %v", f.branch, err)
		return nil, err
	}

	ref, err := r.Head()
	if err != nil {
		return nil, err
	}

	commit, err := r.CommitObject(ref.Hash())
	if err != nil {
		return nil, err
	}

	log.Infof("HEAD ref hash '%s'", ref.Hash())

	return commit, nil
}

// NewAuth creates new AuthMethod based on URI
func NewAuth(git string, sshkey string) (transport.AuthMethod, error) {
	var auth transport.AuthMethod

	ep, err := transport.NewEndpoint(git)
	if err != nil {
		return nil, err
	}

	if strings.HasPrefix(ep.Protocol, "ssh") && sshkey != "" {
		var signer ssh.Signer
		sshFile, err := os.Open(sshkey)
		if err != nil {
			return nil, errors.New("Couldn't open SSH key: " + err.Error())
		}
		sshB, err := ioutil.ReadAll(sshFile)
		if err != nil {
			return nil, errors.New("Couldn't read SSH key: " + err.Error())
		}

		signer, err = ssh.ParsePrivateKey(sshB)
		if err != nil {
			return nil, errors.New("Couldn't parse SSH key: " + err.Error())
		}

		sshAuth := &gitssh.PublicKeys{User: "git", Signer: signer}
		return sshAuth, nil
	}

	if strings.HasPrefix(ep.Protocol, "http") && ep.User != "" && ep.Password != "" {
		auth = &githttp.BasicAuth{
			Username: ep.User,
			Password: ep.Password,
		}
	}

	return auth, nil
}
