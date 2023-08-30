package utils

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage/memory"
)

func CloneGitRepo(repoUrl string, accessToken string, branch string) (*git.Repository, error) {
	r, err := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{
		URL: repoUrl,
		Auth: &http.BasicAuth{
			Username: "x-access-token", // anything except an empty string
			Password: accessToken,
		},
		ReferenceName: plumbing.NewBranchReferenceName(branch),
		SingleBranch:  true,
		Depth:         1,
	})
	if err != nil {
		return nil, err
	}
	return r, nil
}
