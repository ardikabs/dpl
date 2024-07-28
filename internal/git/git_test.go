package git_test

import (
	"context"
	"os"
	"testing"

	"github.com/ardikabs/dpl/internal/git"
	"github.com/ardikabs/dpl/internal/types"
	fixtures "github.com/go-git/go-git-fixtures/v4"
	gogit "github.com/go-git/go-git/v5"

	"github.com/stretchr/testify/require"
)

func getBasicRepositoryURL() string {
	fixture := fixtures.Basic().One()
	return getRepositoryURL(fixture)
}

func getRepositoryURL(f *fixtures.Fixture) string {
	return f.DotGit().Root()
}

func getTempDir(t *testing.T) string {
	tmp, err := os.MkdirTemp("/tmp", "git-test-*")
	require.NoError(t, err)

	t.Cleanup(func() {
		os.RemoveAll(tmp)
	})

	return tmp
}

func TestGit_Clone(t *testing.T) {
	t.Run("normal clone", func(t *testing.T) {
		g, err := git.New(types.GitSecret{})
		require.NoError(t, err)

		dir := getTempDir(t)
		repo, err := g.Clone(context.TODO(), getBasicRepositoryURL(), dir)
		require.NoError(t, err)
		require.NotNil(t, repo)
	})

	t.Run("if repository is already exists, no error returned", func(t *testing.T) {
		dir := getTempDir(t)
		_, err := gogit.PlainClone(dir, false, &gogit.CloneOptions{
			URL: getBasicRepositoryURL(),
		})
		require.NoError(t, err)

		g, err := git.New(types.GitSecret{})
		require.NoError(t, err)

		repo, err := g.Clone(context.TODO(), getBasicRepositoryURL(), dir)
		require.NoError(t, err)
		require.NotNil(t, repo)
	})
}
