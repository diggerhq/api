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
		TriggerTestJob(client, repoOwner, repoName, &job, workflowFileName)
	}
	return nil
}

func TriggerTestJob(client *github.Client, repoOwner string, repoName string, job *models.DiggerJob, workflowFileName string) {
	log.Printf("TriggerTestJob jobId: %v", job.DiggerJobId)
	ctx := context.Background()
	if job.SerializedJob == nil {
		log.Printf("GitHub job can't be nil")
	}
	jobString := string(job.SerializedJob)
	log.Printf("jobString: %v \n", jobString)
	_, err := client.Actions.CreateWorkflowDispatchEventByFileName(ctx, repoOwner, repoName, workflowFileName, github.CreateWorkflowDispatchEventRequest{
		Ref:    job.BranchName,
		Inputs: map[string]interface{}{"job": jobString, "id": job.DiggerJobId},
	})
	if err != nil {
		log.Printf("TriggerTestJob err: %v\n", err)
		return
	}
}
