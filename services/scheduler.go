package services

import (
	"context"
	"digger.dev/cloud/models"
	"fmt"
	"github.com/google/go-github/v54/github"
)

func DiggerJobCompleted(client *github.Client, jobId string, repoOwner string, repoName string, workflowFileName string) error {
	job, err := models.GetDiggerJobByParentId(jobId)
	if err != nil {
		return err
	}

	diggerJobId := job.DiggerJobId
	TriggerTestJob(client, repoOwner, repoName, diggerJobId, workflowFileName)
	return nil
}

func TriggerTestJob(client *github.Client, repoOwner string, repoName string, jobId string, workflowFileName string) {
	//_, _, _ := client.Repositories.Get(ctx, owner, repo_name)
	ctx := context.Background()
	event := github.CreateWorkflowDispatchEventRequest{Ref: "main", Inputs: map[string]interface{}{"id": jobId}}
	_, err := client.Actions.CreateWorkflowDispatchEventByFileName(ctx, repoOwner, repoName, workflowFileName, event)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}
}
