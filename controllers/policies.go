package controllers

import (
	"context"
	"digger.dev/cloud/middleware"
	"digger.dev/cloud/models"
	"errors"
	"fmt"
	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v54/github"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"io"
	"log"
	"net/http"
	"os"
)

type CreatePolicyInput struct {
	Policy string
}

func FindAccessPolicy(c *gin.Context) {
	findPolicy(c, models.POLICY_TYPE_ACCESS)
}

func FindPlanPolicy(c *gin.Context) {
	findPolicy(c, models.POLICY_TYPE_PLAN)
}

func FindDriftPolicy(c *gin.Context) {
	findPolicy(c, models.POLICY_TYPE_DRIFT)
}

func findPolicy(c *gin.Context, policyType string) {
	namespace := c.Param("namespace")
	projectName := c.Param("projectName")
	orgId, exists := c.Get(middleware.ORGANISATION_ID_KEY)

	if !exists {
		log.Printf("Organisation ID not found in context")
		c.String(http.StatusForbidden, "Not allowed to access this resource")
		return
	}

	var policy models.Policy
	query := JoinedOrganisationNamespaceProjectQuery()

	if namespace != "" && projectName != "" {
		err := query.
			Where("namespaces.name = ? AND projects.name = ? AND policies.organisation_id = ? AND policies.type = ?", namespace, projectName, orgId, policyType).
			First(&policy).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.String(http.StatusNotFound, fmt.Sprintf("Could not find policy for namespace %v and project name %v", namespace, projectName))
			} else {
				c.String(http.StatusInternalServerError, "Unknown error occurred while fetching database")
			}
			return
		}
	} else {
		c.String(http.StatusBadRequest, "Should pass namespace and project name")
		return
	}

	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.String(http.StatusOK, policy.Policy)
}

func FindAccessPolicyForOrg(c *gin.Context) {
	findPolicyForOrg(c, models.POLICY_TYPE_ACCESS)
}

func FindPlanPolicyForOrg(c *gin.Context) {
	findPolicyForOrg(c, models.POLICY_TYPE_PLAN)
}

func FindDriftPolicyForOrg(c *gin.Context) {
	findPolicyForOrg(c, models.POLICY_TYPE_DRIFT)
}

func findPolicyForOrg(c *gin.Context, policyType string) {
	organisation := c.Param("organisation")
	var policy models.Policy
	query := JoinedOrganisationNamespaceProjectQuery()

	err := query.
		Where("organisations.name = ? AND (namespaces.id IS NULL AND projects.id IS NULL) AND policies.type = ? ", organisation, policyType).
		First(&policy).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.String(http.StatusNotFound, "Could not find policy for organisation: "+organisation)
		} else {
			c.String(http.StatusInternalServerError, "Unknown error occurred while fetching database")
		}
		return
	}

	loggedInOrganisation := c.GetUint(middleware.ORGANISATION_ID_KEY)

	if policy.OrganisationID != loggedInOrganisation {
		log.Printf("Organisation ID %v does not match logged in organisation ID %v", policy.OrganisationID, loggedInOrganisation)
		c.String(http.StatusForbidden, "Not allowed to access this resource")
		return
	}

	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.String(http.StatusOK, policy.Policy)
}

func JoinedOrganisationNamespaceProjectQuery() *gorm.DB {
	return models.DB.Preload("Organisation").Preload("Namespace").Preload("Project").
		Joins("LEFT JOIN namespaces ON policies.namespace_id = namespaces.id").
		Joins("LEFT JOIN projects ON policies.project_id = projects.id").
		Joins("LEFT JOIN organisations ON policies.organisation_id = organisations.id")
}

func UpsertAccessPolicyForOrg(c *gin.Context) {
	upsertPolicyForOrg(c, models.POLICY_TYPE_ACCESS)
}

func UpsertPlanPolicyForOrg(c *gin.Context) {
	upsertPolicyForOrg(c, models.POLICY_TYPE_PLAN)
}

func UpsertDriftPolicyForOrg(c *gin.Context) {
	upsertPolicyForOrg(c, models.POLICY_TYPE_DRIFT)
}

func upsertPolicyForOrg(c *gin.Context, policyType string) {
	// Validate input
	policyData, err := io.ReadAll(c.Request.Body)
	if err != nil {
		// Handle the error
		c.String(http.StatusInternalServerError, "Error reading request body")
		return
	}
	organisation := c.Param("organisation")

	org := models.Organisation{}
	orgResult := models.DB.Where("name = ?", organisation).Take(&org)
	if orgResult.RowsAffected == 0 {
		c.String(http.StatusNotFound, "Could not find organisation: "+organisation)
		return
	}

	loggedInOrganisation := c.GetUint(middleware.ORGANISATION_ID_KEY)

	if org.ID != loggedInOrganisation {
		log.Printf("Organisation ID %v does not match logged in organisation ID %v", org.ID, loggedInOrganisation)
		c.String(http.StatusForbidden, "Not allowed to access this resource")
		return
	}

	policy := models.Policy{}

	policyResult := models.DB.Where("organisation_id = ? AND (namespace_id IS NULL AND project_id IS NULL) AND type = ?", org.ID, policyType).Take(&policy)

	if policyResult.RowsAffected == 0 {
		err := models.DB.Create(&models.Policy{
			OrganisationID: org.ID,
			Type:           policyType,
			Policy:         string(policyData),
		}).Error

		if err != nil {
			log.Printf("Error creating policy: %v", err)
			c.String(http.StatusInternalServerError, "Error creating policy")
			return
		}
	} else {
		err := policyResult.Update("policy", string(policyData)).Error
		if err != nil {
			log.Printf("Error updating policy: %v", err)
			c.String(http.StatusInternalServerError, "Error updating policy")
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func UpsertAccessPolicyForNamespaceAndProject(c *gin.Context) {
	upsertPolicyForNamespaceAndProject(c, models.POLICY_TYPE_ACCESS)
}

func UpsertPlanPolicyForNamespaceAndProject(c *gin.Context) {
	upsertPolicyForNamespaceAndProject(c, models.POLICY_TYPE_PLAN)
}

func UpsertDriftPolicyForNamespaceAndProject(c *gin.Context) {
	upsertPolicyForNamespaceAndProject(c, models.POLICY_TYPE_DRIFT)
}

func upsertPolicyForNamespaceAndProject(c *gin.Context, policyType string) {
	orgID, exists := c.Get(middleware.ORGANISATION_ID_KEY)

	if !exists {
		c.String(http.StatusUnauthorized, "Not authorized")
		return
	}

	orgID = orgID.(uint)

	// Validate input
	policyData, err := io.ReadAll(c.Request.Body)
	if err != nil {
		// Handle the error
		c.String(http.StatusInternalServerError, "Error reading request body")
		return
	}
	namespace := c.Param("namespace")
	projectName := c.Param("projectName")
	namespaceModel := models.Namespace{}
	namespaceResult := models.DB.Where("name = ?", namespace).Take(&namespaceModel)
	if namespaceResult.RowsAffected == 0 {
		namespaceModel = models.Namespace{
			OrganisationID: orgID.(uint),
			Name:           namespace,
		}
		result := models.DB.Create(&namespaceModel)
		if result.Error != nil {
			log.Printf("Error creating namespace: %v", err)
			c.String(http.StatusInternalServerError, "Error creating missing namespace")
			return
		}
	}

	projectModel := models.Project{}
	projectResult := models.DB.Where("name = ?", projectName).Take(&projectModel)
	if projectResult.RowsAffected == 0 {
		projectModel = models.Project{
			OrganisationID: orgID.(uint),
			NamespaceID:    namespaceModel.ID,
			Name:           projectName,
		}
		err := models.DB.Create(&projectModel).Error
		if err != nil {
			log.Printf("Error creating project: %v", err)
			c.String(http.StatusInternalServerError, "Error creating missing project")
			return
		}
	}

	var policy models.Policy

	policyResult := models.DB.Where("organisation_id = ? AND namespace_id = ? AND project_id = ? AND type = ?", orgID, namespaceModel.ID, projectModel.ID, policyType).Take(&policy)

	if policyResult.RowsAffected == 0 {
		err := models.DB.Create(&models.Policy{
			OrganisationID: orgID.(uint),
			NamespaceID:    &namespaceModel.ID,
			ProjectID:      &projectModel.ID,
			Type:           policyType,
			Policy:         string(policyData),
		}).Error
		if err != nil {
			log.Printf("Error creating policy: %v", err)
			c.String(http.StatusInternalServerError, "Error creating policy")
			return
		}
	} else {
		err := policyResult.Update("policy", string(policyData)).Error
		if err != nil {
			log.Printf("Error updating policy: %v", err)
			c.String(http.StatusInternalServerError, "Error updating policy")
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func IssueAccessTokenForOrg(c *gin.Context) {
	organisation_ID, exists := c.Get(middleware.ORGANISATION_ID_KEY)

	if !exists {
		c.String(http.StatusUnauthorized, "Not authorized")
		return
	}

	org := models.Organisation{}
	orgResult := models.DB.Where("id = ?", organisation_ID).Take(&org)
	if orgResult.RowsAffected == 0 {
		log.Printf("Could not find organisation: %v", organisation_ID)
		c.String(http.StatusInternalServerError, "Unexpected error")
		return
	}

	// prefixing token to make easier to retire this type of tokens later
	token := "t:" + uuid.New().String()

	err := models.DB.Create(&models.Token{
		Value:          token,
		OrganisationID: org.ID,
		Type:           models.AccessPolicyType,
	}).Error

	if err != nil {
		log.Printf("Error creating token: %v", err)
		c.String(http.StatusInternalServerError, "Unexpected error")
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

func GithubWebhookHandler(c *gin.Context) {
	payload, err := github.ValidatePayload(c.Request, []byte(os.Getenv("GITHUB_WEBHOOK_SECRET")))
	if err != nil {
		log.Printf("Error validating payload: %v", err)
		c.String(http.StatusBadRequest, "Error validating payload")
		return
	}

	event, err := github.ParseWebHook(github.WebHookType(c.Request), payload)
	if err != nil {
		log.Printf("Error parsing webhook: %v", err)
		c.String(http.StatusBadRequest, "Error parsing webhook")
		return
	}

	switch event := event.(type) {
	case *github.InstallationEvent:
		log.Printf("Got installation event for %v", event.GetInstallation().GetAccount().GetLogin())
		if event.GetAction() == "created" {
			err := models.DB.Create(&models.GithubAppInstallation{
				GithubInstallationId: *event.Installation.ID,
				GithubAppId:          *event.Installation.AppID,
			}).Error
			if err != nil {
				log.Printf("Error creating github event: %v", err)
			}
			c.String(http.StatusOK, "OK")
		}
	case *github.InstallationRepositoriesEvent:
		log.Printf("Got installation event for %v", event.GetInstallation().GetAccount().GetLogin())
		if event.GetAction() == "added" {
			err := models.DB.Create(&models.GithubAppInstallation{
				GithubInstallationId: *event.Installation.ID,
				GithubAppId:          *event.Installation.AppID,
			}).Error
			if err != nil {
				log.Printf("Error creating github event: %v", err)
			}
			c.String(http.StatusOK, "OK")
		}
	case *github.PullRequestEvent:
		log.Printf("Got pull request event for %v", event.GetPullRequest().GetTitle())
		installation := models.GithubAppInstallation{}
		err := models.DB.Where("github_installation_id = ?", *event.Installation.ID).Take(&installation).Error
		if err != nil {
			log.Printf("Error getting installation: %v", err)
			c.String(http.StatusInternalServerError, "Error getting installation")
			return
		}
		ghApp := models.GithubApp{}
		err = models.DB.Where("github_id = ?", installation.GithubAppId).Take(&ghApp).Error
		if err != nil {
			log.Printf("Error getting app: %v", err)
			c.String(http.StatusInternalServerError, "Error getting app")
			return
		}
		tr := http.DefaultTransport

		itr, err := ghinstallation.New(tr, installation.GithubAppId, installation.GithubInstallationId, []byte(ghApp.PrivateKey))
		if err != nil {
			log.Printf("Error initialising installation: %v", err)
			c.String(http.StatusInternalServerError, "Error getting app")
			return
		}

		client := github.NewClient(&http.Client{Transport: itr})

		jobString := "[\n  {\n    \"projectName\": \"prod\",\n    \"projectDir\": \"prod\",\n    \"projectWorkspace\": \"default\",\n    \"terragrunt\": false,\n    \"commands\": [\"digger plan\"],\n    \"applyStage\": {\n      \"steps\": [\n        { \"action\": \"init\", \"extraArgs\": [] },\n        { \"action\": \"apply\", \"extraArgs\": [] }\n      ]\n    },\n    \"planStage\": {\n      \"steps\": [\n        { \"action\": \"init\", \"extraArgs\": [] },\n        { \"action\": \"plan\", \"extraArgs\": [] }\n      ]\n    },\n    \"pullRequestNumber\": 1,\n    \"eventName\": \"pull_request\",\n    \"requestedBy\": \"Spartakovic\",\n    \"namespace\": \"diggerhq/digger-demo-ghapp\",\n    \"stateEnvVars\": {\n      \"TF_VAR_aws_region\": \"us-east-1\",\n      \"TF_VAR_aws_access_key_id\": \"AKIA\"\n    },\n    \"commandEnvVars\": {\n      \"TF_VAR_aws_region\": \"us-east-1\",\n      \"TF_VAR_aws_access_key_id\": \"AKIA\"\n    }\n  }\n]\n"

		resp, err := client.Actions.CreateWorkflowDispatchEventByFileName(context.Background(), *event.Organization.Login, *event.Repo.Name, "plan.yml", github.CreateWorkflowDispatchEventRequest{
			Ref:    event.PullRequest.Head.GetRef(),
			Inputs: map[string]interface{}{"jobs": jobString},
		})

		if err != nil {
			log.Printf("Error getting app: %v", err)
			c.String(http.StatusInternalServerError, "Error getting app")
			return
		}

		log.Printf("Got response: %v", resp)

	default:
		log.Printf("Unhandled event type: %v", event)
	}

	c.String(http.StatusOK, "OK")
}

type Step struct {
	Action    string   `json:"action"`
	ExtraArgs []string `json:"extraArgs"`
}

type Stage struct {
	Steps []Step `json:"steps"`
}

type Job struct {
	ProjectName       string            `json:"projectName"`
	ProjectDir        string            `json:"projectDir"`
	ProjectWorkspace  string            `json:"projectWorkspace"`
	Terragrunt        bool              `json:"terragrunt"`
	Commands          []string          `json:"commands"`
	ApplyStage        Stage             `json:"applyStage"`
	PlanStage         Stage             `json:"planStage"`
	PullRequestNumber int               `json:"pullRequestNumber"`
	EventName         string            `json:"eventName"`
	RequestedBy       string            `json:"requestedBy"`
	Namespace         string            `json:"namespace"`
	StateEnvVars      map[string]string `json:"stateEnvVars"`
	CommandEnvVars    map[string]string `json:"commandEnvVars"`
}
