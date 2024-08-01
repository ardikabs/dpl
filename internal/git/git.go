package git

import (
	"errors"

	"github.com/ardikabs/dpl/internal/types"
	"github.com/go-git/go-git/v5"
	"golang.org/x/net/context"
)

var (
	RemoteName = "origin"

	ErrGitAccessFailed = errors.New("git access is failed to initialize")
)

type Git struct {
	secret types.GitSecret
}

func New(secret types.GitSecret) (*Git, error) {
	g := &Git{
		secret: secret,
	}
	return g, nil
}

func (g *Git) Clone(ctx context.Context, url, dest string, opts ...CloneOption) (Repository, error) {
	o := NewDefaultCloneOptions()
	for _, opt := range opts {
		opt(o)
	}

	authMethod := g.secret.GetAuthMethod()

	gitRepo, err := git.PlainClone(dest, false, &git.CloneOptions{
		URL:           url,
		Auth:          authMethod,
		SingleBranch:  o.SingleBranch,
		ReferenceName: o.Reference,
		RemoteName:    RemoteName,
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

	r := NewGitRepository(gitRepo, authMethod)
	return r, nil
}
