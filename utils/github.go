package utils

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"log"
	"os"
)

func createTempDir() string {
	tempDir, err := os.MkdirTemp("", "repo")
	if err != nil {
		log.Fatal(err)
	}
	return tempDir
}

type action func(string)

func CloneGitRepoAndDoAction(repoUrl string, branch string, accessToken string, action action) error {
	dir := createTempDir()
	_, err := git.PlainClone(dir, false, &git.CloneOptions{
		URL: repoUrl,
		Auth: &http.BasicAuth{
			Username: "x-access-token", // anything except an empty string
			Password: accessToken,
		},
		ReferenceName: plumbing.NewBranchReferenceName(branch),
		Depth:         1,
		SingleBranch:  true,
	})
	if err != nil {
		return err
	}

	action(dir)

	defer os.RemoveAll(dir)
	return nil

}
