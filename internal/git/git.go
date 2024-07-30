package git

import (
	"errors"
	"fmt"
	"strings"

	"github.com/ardikabs/dpl/internal/types"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"golang.org/x/net/context"
)

var (
	ErrGitAccessFailed = errors.New("git access is failed to initialize")

	RemoteName = "origin"
)

type Git struct {
	rawAuth string
	auth    transport.AuthMethod
}

func New(secret types.GitSecret) (*Git, error) {
	g := &Git{
		rawAuth: secret.Raw(),
		auth:    &http.BasicAuth{Username: secret.Username, Password: secret.Password},
	}
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

	if err := g.setGitRepoConfig(gitRepo); err != nil {
		return nil, err
	}

	r := NewGitRepository(gitRepo, g.auth)
	return r, nil
}

func (g *Git) setGitRepoConfig(repo *git.Repository) error {
	cfg, err := repo.Config()
	if err != nil {
		return err
	}

	for _, remote := range cfg.Remotes {
		if remote.Name == RemoteName && len(remote.URLs) > 0 {
			url := remote.URLs[0]
			urlParts := strings.Split(url, "//")
			remote.URLs = []string{fmt.Sprintf("%s//%s@%s", urlParts[0], g.rawAuth, urlParts[1])}
		}
	}

	for _, branch := range cfg.Branches {
		branch.Rebase = "true"
	}

	if err := repo.SetConfig(cfg); err != nil {
		return err
	}

	return nil
}
