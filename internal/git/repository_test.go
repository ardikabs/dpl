package git_test

import (
	"testing"

	"github.com/ardikabs/dpl/internal/git"
	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getDummyRepoAuth() *http.BasicAuth {
	return &http.BasicAuth{}
}

func TestRepository_Root(t *testing.T) {
	destDir := getTempDir(t)
	gitRepo, err := gogit.PlainClone(destDir, false, &gogit.CloneOptions{
		URL: getBasicRepositoryURL(),
	})
	require.NoError(t, err)

	r, err := git.NewGitRepository(gitRepo, getDummyRepoAuth())
	assert.NoError(t, err)
	assert.Equal(t, r.Root(), destDir)
}
