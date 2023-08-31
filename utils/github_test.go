package utils

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestGitHubCloneWithInvalidTokenThrowsErr(t *testing.T) {
	f := func(d string) {}
	err := CloneGitRepoAndDoAction("https://github.com/diggerhq/private-repo", "main", "invalid-token", f)
	assert.NotNil(t, err)
}

func TestGitHubCloneWithPublicRepoThrowsNoError(t *testing.T) {
	token := os.Getenv("GITHUB_PAT_TOKEN")
	f := func(d string) {}
	err := CloneGitRepoAndDoAction("https://github.com/diggerhq/digger", "develop", token, f)
	assert.Nil(t, err)
}

func TestGitHubCloneWithInvalidBranchThrowsError(t *testing.T) {
	token := os.Getenv("GITHUB_PAT_TOKEN")
	f := func(d string) {}
	err := CloneGitRepoAndDoAction("https://github.com/diggerhq/digger", "not-a-branch", token, f)
	assert.NotNil(t, err)
}