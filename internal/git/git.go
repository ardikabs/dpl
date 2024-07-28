package git

import (
	"errors"

	"github.com/ardikabs/dpl/internal/types"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"golang.org/x/net/context"
)

var (
	ErrGitAccessFailed = errors.New("git access is failed to initialize")
)

type Git struct {
	auth transport.AuthMethod
}

func New(secret types.GitSecret) (*Git, error) {
	g := &Git{auth: &http.BasicAuth{Username: secret.Username, Password: secret.Password}}
	return g, nil
}

func (g *Git) Clone(ctx context.Context, url, dest string, opts ...CloneOption) (Repository, error) {
	o := NewDefaultCloneOptions()
	for _, opt := range opts {
		opt(o)
	}

	gitRepo, err := git.PlainClone(dest, false, &git.CloneOptions{
		URL:           url,
		Auth:          g.auth,
		SingleBranch:  o.SingleBranch,
		ReferenceName: o.Reference,
	})
	if err != nil {
		if errors.Is(err, git.ErrRepositoryAlreadyExists) {
			if gitRepo, err = git.PlainOpen(dest); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	r := NewGitRepository(gitRepo, g.auth)
	return r, nil
}
