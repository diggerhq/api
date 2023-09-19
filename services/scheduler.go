package services

import (
	"context"
	"digger.dev/cloud/models"
	"github.com/google/go-github/v55/github"
	"log"
)

func DiggerJobCompleted(client *github.Client, parentJob *models.DiggerJob, repoOwner string, repoName string, workflowFileName string) error {
	log.Printf("DiggerJobCompleted parentJobId: %v", parentJob.DiggerJobId)

	jobs, err := models.DB.GetDiggerJobsByParentIdAndStatus(&parentJob.DiggerJobId, models.DiggerJobCreated)
	if err != nil {
		return err
	}

	for _, job := range jobs {
		TriggerJob(client, repoOwner, repoName, &job, workflowFileName)
	}
	return nil
}

func TriggerJob(client *github.Client, repoOwner string, repoName string, job *models.DiggerJob, workflowFileName string) {
	log.Printf("TriggerJob jobId: %v", job.DiggerJobId)
	ctx := context.Background()
	if job.SerializedJob == nil {
		log.Printf("GitHub job can't me nil")
	}
	jobString := string(job.SerializedJob)
	log.Printf("jobString: %v \n", jobString)
	_, err := client.Actions.CreateWorkflowDispatchEventByFileName(ctx, repoOwner, repoName, workflowFileName, github.CreateWorkflowDispatchEventRequest{
		Ref:    job.BranchName,
		Inputs: map[string]interface{}{"job": jobString, "id": job.DiggerJobId},
	})
	if err != nil {
		log.Printf("TriggerJob err: %v\n", err)
		return
	}
}
