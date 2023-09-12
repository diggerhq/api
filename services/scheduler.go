package services

import (
	"context"
	"digger.dev/cloud/models"
	"fmt"
	"github.com/google/go-github/v55/github"
)

func DiggerJobCompleted(client *github.Client, parentJob *models.DiggerJob, repoOwner string, repoName string, workflowFileName string) error {
	jobs, err := models.DB.GetDiggerJobsByParentId(&parentJob.DiggerJobId)
	if err != nil {
		return err
	}

	for _, job := range jobs {
		TriggerTestJob(client, repoOwner, repoName, &job, workflowFileName)
	}
	return nil
}

func TriggerTestJob(client *github.Client, repoOwner string, repoName string, job *models.DiggerJob, workflowFileName string) {
	//_, _, _ := client.Repositories.Get(ctx, owner, repo_name)
	ctx := context.Background()
	event := github.CreateWorkflowDispatchEventRequest{Ref: "main", Inputs: map[string]interface{}{"id": job.DiggerJobId}}
	_, err := client.Actions.CreateWorkflowDispatchEventByFileName(ctx, repoOwner, repoName, workflowFileName, event)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}
}
