package main

import (
	"context"
	"os/exec"
	"testing"

	"github.com/Masterminds/semver/v3"

	"github.com/stretchr/testify/require"
)

func TestFoo(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	c := exec.Command("sh", "-c", `
git init
git commit --allow-empty -m "first"
git tag v0.1.0
git tag foo0.1.0
git tag v0.1.1
git tag foo0.1.1
git tag v0.2.0
git tag foo0.2.0
git tag v1.0.0
git commit --allow-empty -m "second"
git commit --allow-empty -m "third"
git tag v2.0.0
git tag foo
git commit --allow-empty -m "forth"
git tag bar
`,
	)
	c.Dir = dir
	require.NoError(t, c.Run())

	t.Run("", func(t *testing.T) {
		opts := prevVersionOptions{
			repoDir:  dir,
			prefixes: []string{"v"},
		}
		got, err := prevVersion(ctx, &opts)
		require.NoError(t, err)
		require.Equal(t, "v2.0.0", got)
	})

	t.Run("", func(t *testing.T) {
		opts := prevVersionOptions{
			repoDir:  dir,
			prefixes: []string{"v", "bar"},
		}
		got, err := prevVersion(ctx, &opts)
		require.NoError(t, err)
		require.Equal(t, "v2.0.0", got)
	})

	t.Run("", func(t *testing.T) {
		versionZero, err := semver.NewConstraint("< 1.0.0")
		require.NoError(t, err)
		opts := prevVersionOptions{
			repoDir:     dir,
			prefixes:    []string{"v", "foo"},
			constraints: versionZero,
		}
		got, err := prevVersion(ctx, &opts)
		require.NoError(t, err)
		require.Equal(t, "v0.2.0", got)
	})
}
