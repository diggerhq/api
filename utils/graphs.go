package utils

import (
	"digger.dev/cloud/models"
	"encoding/json"
	"fmt"
	"github.com/diggerhq/lib-digger-config"
	"github.com/diggerhq/lib-orchestrator"
	"github.com/dominikbraun/graph"
	"github.com/google/uuid"
	"log"
)

// ConvertJobsToDiggerJobs jobs is map with project name as a key and a Job as a value
func ConvertJobsToDiggerJobs(jobsMap map[string]orchestrator.Job, projectMap map[string]configuration.Project, projectsGraph graph.Graph[string, configuration.Project], branch string, repoFullName string) (*uuid.UUID, map[string]*models.DiggerJob, error) {
	result := make(map[string]*models.DiggerJob)

	log.Printf("Number of Jobs: %v\n", len(jobsMap))
	marshalledJobsMap := map[string][]byte{}
	for _, job := range jobsMap {
		marshalled, _ := json.Marshal(orchestrator.JobToJson(job))
		marshalledJobsMap[job.ProjectName] = marshalled
	}

	batchId, _ := uuid.NewUUID()

	graphWithImpactedProjectsOnly, err := ImpactedProjectsOnlyGraph(projectsGraph, projectMap)

	if err != nil {
		return nil, nil, err
	}

	predecessorMap, err := graphWithImpactedProjectsOnly.PredecessorMap()

	if err != nil {
		return nil, nil, err
	}
	visit := func(value string) bool {
		if predecessorMap[value] == nil || len(predecessorMap[value]) == 0 {
			fmt.Printf("no parent for %v\n", value)
			parentJob, err := models.DB.CreateDiggerJob(batchId, nil, marshalledJobsMap[value], branch)
			if err != nil {
				log.Printf("failed to create a job")
				return false
			}
			_, err = models.DB.CreateDiggerJobLink(parentJob.DiggerJobId, repoFullName)
			if err != nil {
				log.Printf("failed to create a digger job link")
				return false
			}
			result[value] = parentJob
			return false
		} else {
			parents := predecessorMap[value]
			for _, edge := range parents {
				parent := edge.Source
				fmt.Printf("parent: %v\n", parent)
				parentDiggerJob := result[parent]
				childJob, err := models.DB.CreateDiggerJob(batchId, &parentDiggerJob.DiggerJobId, marshalledJobsMap[value], branch)
				if err != nil {
					log.Printf("failed to create a job")
					return false
				}
				_, err = models.DB.CreateDiggerJobLink(childJob.DiggerJobId, repoFullName)
				if err != nil {
					log.Printf("failed to create a digger job link")
					return false
				}
				result[value] = childJob
			}
			return false
		}
	}
	err = TraverseGraphVisitAllParentsFirst(graphWithImpactedProjectsOnly, visit)

	if err != nil {
		return nil, nil, err
	}

	return &batchId, result, nil
}

func TraverseGraphVisitAllParentsFirst(graphWithImpactedProjectsOnly graph.Graph[string, configuration.Project], visit func(value string) bool) error {
	dummyParent := configuration.Project{Name: "DUMMY_PARENT_PROJECT_FOR_PROCESSING"}
	predecessorMap, err := graphWithImpactedProjectsOnly.PredecessorMap()
	if err != nil {
		return err
	}

	visitIgnoringDummyParent := func(value string) bool {
		if value == dummyParent.Name {
			return false
		}
		return visit(value)
	}

	err = graphWithImpactedProjectsOnly.AddVertex(dummyParent)
	if err != nil {
		return err
	}
	for node := range predecessorMap {
		if predecessorMap[node] == nil || len(predecessorMap[node]) == 0 {
			err := graphWithImpactedProjectsOnly.AddEdge(dummyParent.Name, node)
			if err != nil {
				return err
			}
		}
	}
	return graph.BFS(graphWithImpactedProjectsOnly, dummyParent.Name, visitIgnoringDummyParent)
}

func ImpactedProjectsOnlyGraph(projectsGraph graph.Graph[string, configuration.Project], projectMap map[string]configuration.Project) (graph.Graph[string, configuration.Project], error) {
	adjacencyMap, err := projectsGraph.AdjacencyMap()
	if err != nil {
		return nil, err
	}
	predecessorMap, err := projectsGraph.PredecessorMap()
	if err != nil {
		return nil, err
	}

	graphWithImpactedProjectsOnly := graph.NewLike(projectsGraph)

	for node := range predecessorMap {
		if _, ok := projectMap[node]; (predecessorMap[node] == nil || len(predecessorMap[node]) == 0) && ok {
			err := CollapsedGraph(nil, node, adjacencyMap, graphWithImpactedProjectsOnly, projectMap)
			if err != nil {
				return nil, err
			}
		}
	}
	return graphWithImpactedProjectsOnly, nil
}

func CollapsedGraph(impactedParent *string, currentNode string, adjMap map[string]map[string]graph.Edge[string], g graph.Graph[string, configuration.Project], impactedProjects map[string]configuration.Project) error {
	// add to the resulting graph only if the project has been impacted by changes
	if _, ok := impactedProjects[currentNode]; ok {
		currentProject, ok := impactedProjects[currentNode]
		if !ok {
			return fmt.Errorf("project %s not found", currentNode)
		}

		err := g.AddVertex(currentProject)
		if err != nil {
			return err
		}
		// process all children nodes
		for child, _ := range adjMap[currentNode] {
			err := CollapsedGraph(&currentNode, child, adjMap, g, impactedProjects)
			if err != nil {
				return err
			}
		}
		// if there is an impacted parent add an edge
		if impactedParent != nil {
			err := g.AddEdge(*impactedParent, currentNode)
			if err != nil {
				return err
			}
		}
	} else {
		// if current wasn't impacted, see children of current node and set currently known parent
		for child, _ := range adjMap[currentNode] {
			err := CollapsedGraph(impactedParent, child, adjMap, g, impactedProjects)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
