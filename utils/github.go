package utils

import (
	"context"
	"fmt"
	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/google/go-github/v55/github"
	"log"
	net "net/http"
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

func CloneGitRepoAndDoAction(repoUrl string, branch string, token string, action action) error {
	dir := createTempDir()
	_, err := git.PlainClone(dir, false, &git.CloneOptions{
		URL: repoUrl,
		Auth: &http.BasicAuth{
			Username: "x-access-token", // anything except an empty string
			Password: token,
		},
		ReferenceName: plumbing.NewBranchReferenceName(branch),
		Depth:         1,
		SingleBranch:  true,
	})
	if err != nil {
		fmt.Printf("PlainClone error: %v\n", err)
		return err
	}

	action(dir)

	defer os.RemoveAll(dir)
	return nil

}

func GetGithubClient(githubAppId int64, installationId int64) (*github.Client, *string, error) {
	githubAppPrivateKey := os.Getenv("GITHUB_APP_PRIVATE_KEY")
	tr := net.DefaultTransport
	itr, err := ghinstallation.New(tr, githubAppId, installationId, []byte(githubAppPrivateKey))
	if err != nil {
		return nil, nil, fmt.Errorf("error initialising github app installation: %v\n", err)
	}

	token, err := itr.Token(context.Background())
	if err != nil {
		return nil, nil, fmt.Errorf("error initialising git app token: %v\n", err)
	}
	ghClient := github.NewClient(&net.Client{Transport: itr})
	return ghClient, &token, nil
}
