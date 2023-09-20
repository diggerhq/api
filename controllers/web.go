package controllers

import (
	"digger.dev/cloud/config"
	"digger.dev/cloud/middleware"
	"digger.dev/cloud/models"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/robert-nix/ansihtml"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type WebController struct {
	Config *config.Config
}

func (web *WebController) validateRequestProjectId(c *gin.Context) (*models.Project, bool) {
	projectId64, err := strconv.ParseUint(c.Param("projectid"), 10, 32)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to parse project id")
		return nil, false
	}
	projectId := uint(projectId64)
	projects, done := models.DB.GetProjectsFromContext(c, middleware.ORGANISATION_ID_KEY)
	if !done {
		return nil, false
	}

	for _, p := range projects {
		if projectId == p.ID {
			return &p, true
		}
	}

	c.String(http.StatusForbidden, "Not allowed to access this resource")
	return nil, false
}

func (web *WebController) ProjectsPage(c *gin.Context) {
	projects, done := models.DB.GetProjectsFromContext(c, middleware.ORGANISATION_ID_KEY)
	if !done {
		return
	}

	c.HTML(http.StatusOK, "projects.tmpl", gin.H{
		"Projects": projects,
	})
}

func (web *WebController) ReposPage(c *gin.Context) {
	repos, done := models.DB.GetReposFromContext(c, middleware.ORGANISATION_ID_KEY)
	if !done {
		return
	}

	githubAppId := os.Getenv("GITHUB_APP_ID")
	githubApp, err := models.DB.GetGithubApp(githubAppId)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to find GitHub app")
		return
	}

	c.HTML(http.StatusOK, "repos.tmpl", gin.H{
		"Repos":     repos,
		"GithubApp": githubApp,
	})
}

func (web *WebController) AddProjectPage(c *gin.Context) {
	if c.Request.Method == "GET" {
		message := ""

		c.HTML(http.StatusOK, "project_add.tmpl", gin.H{
			"Message": message,
		})
	} else if c.Request.Method == "POST" {
		message := ""
		repo, ok := models.DB.GetDefaultRepo(c, middleware.ORGANISATION_ID_KEY)
		if !ok {
			message = "failed to create a new project"
			c.HTML(http.StatusOK, "project_add.tmpl", gin.H{
				"Message": message,
			})
		}
		projectName := c.PostForm("project_name")
		if projectName == "" {
			message := "Project's name can't be empty"
			c.HTML(http.StatusOK, "project_add.tmpl", gin.H{
				"Message": message,
			})
		}

		//TODO: gorm is trying to insert a new repo and organisation on every insert of a new project,
		// there should be a way to avoid it
		_, err := models.DB.CreateProject(projectName, repo.Organisation, repo)
		if err != nil {
			message := "Failed to create a project"
			c.HTML(http.StatusOK, "project_add.tmpl", gin.H{
				"Message": message,
			})
		}
		c.Redirect(http.StatusFound, "/projects")
	}
}

func (web *WebController) RunsPage(c *gin.Context) {
	runs, done := models.DB.GetProjectRunsFromContext(c, middleware.ORGANISATION_ID_KEY)
	if !done {
		return
	}
	c.HTML(http.StatusOK, "runs.tmpl", gin.H{
		"Runs": runs,
	})
}

func (web *WebController) PoliciesPage(c *gin.Context) {
	policies, done := models.DB.GetPoliciesFromContext(c, middleware.ORGANISATION_ID_KEY)
	if !done {
		return
	}
	log.Println("policies.tmpl")
	c.HTML(http.StatusOK, "policies.tmpl", gin.H{
		"Policies": policies,
	})
}

func (web *WebController) AddPolicyPage(c *gin.Context) {
	if c.Request.Method == "GET" {
		message := ""
		projects, done := models.DB.GetProjectsFromContext(c, middleware.ORGANISATION_ID_KEY)
		if !done {
			return
		}

		policyTypes := make([]string, 0)
		policyTypes = append(policyTypes, "drift")
		policyTypes = append(policyTypes, "terraform")
		policyTypes = append(policyTypes, "access")

		log.Printf("projects: %v\n", projects)

		c.HTML(http.StatusOK, "policy_add.tmpl", gin.H{
			"Message": message, "Projects": projects, "PolicyTypes": policyTypes,
		})
	} else if c.Request.Method == "POST" {
		policyText := c.PostForm("policytext")
		if policyText == "" {
			message := "Policy can't be empty"
			c.HTML(http.StatusOK, "policy_add.tmpl", gin.H{
				"Message": message,
			})
		}

		policyType := c.PostForm("policytype")
		projectIdStr := c.PostForm("projectid")
		projectId64, err := strconv.ParseUint(projectIdStr, 10, 32)
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to parse policy id")
			return
		}
		projectId := uint(projectId64)
		project, ok := models.DB.GetProjectByProjectId(c, projectId, middleware.ORGANISATION_ID_KEY)
		if !ok {
			log.Printf("Failed to fetch specified project by id: %v, %v\n", projectIdStr, err)
			message := "Failed to create a policy"
			c.HTML(http.StatusOK, "policy_add.tmpl", gin.H{
				"Message": message,
			})
		}

		log.Printf("repo: %v\n", project.Repo)

		policy := models.Policy{Project: project, Policy: policyText, Type: policyType, Organisation: project.Organisation, Repo: project.Repo}

		err = models.DB.GormDB.Create(&policy).Error
		if err != nil {
			log.Printf("Failed to create a new policy, %v\n", err)
			message := "Failed to create a policy"
			c.HTML(http.StatusOK, "policy_add.tmpl", gin.H{
				"Message": message,
			})
		}

		c.Redirect(http.StatusFound, "/policies")
	}
}

func (web *WebController) PolicyDetailsPage(c *gin.Context) {
	policyId64, err := strconv.ParseUint(c.Param("policyid"), 10, 32)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to parse policy id")
		return
	}
	policyId := uint(policyId64)
	policy, ok := models.DB.GetPolicyByPolicyId(c, policyId, middleware.ORGANISATION_ID_KEY)
	if !ok {
		return
	}

	log.Println("policy_details.tmpl")
	c.HTML(http.StatusOK, "policy_details.tmpl", gin.H{
		"Policy": policy,
	})
}

func (web *WebController) ProjectDetailsPage(c *gin.Context) {
	project, ok := web.validateRequestProjectId(c)
	if !ok {
		return
	}

	log.Println("project_details.tmpl")
	c.HTML(http.StatusOK, "project_details.tmpl", gin.H{
		"Project": project,
	})
}

func (web *WebController) RunDetailsPage(c *gin.Context) {
	runId64, err := strconv.ParseUint(c.Param("runid"), 10, 32)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to parse project run id")
		return
	}
	runId := uint(runId64)
	run, ok := models.DB.GetProjectByRunId(c, runId, middleware.ORGANISATION_ID_KEY)
	if !ok {
		return
	}

	stateSyncOutput := ""
	terraformPlanOutput := ""
	runOutput := string(ansihtml.ConvertToHTMLWithClasses([]byte(run.Output), "terraform-output-", true))
	runOutput = strings.Replace(runOutput, "  ", "&nbsp;&nbsp;", -1)
	runOutput = strings.Replace(runOutput, "\n", "<br>\n", -1)

	planIndex := strings.Index(runOutput, "Terraform used the selected providers to generate the following execution")
	if planIndex != -1 {
		stateSyncOutput = runOutput[:planIndex]
		terraformPlanOutput = runOutput[planIndex:]

		c.HTML(http.StatusOK, "run_details.tmpl", gin.H{
			"Run":                      run,
			"TerraformStateSyncOutput": template.HTML(stateSyncOutput),
			"TerraformPlanOutput":      template.HTML(terraformPlanOutput),
		})
	} else {
		c.HTML(http.StatusOK, "run_details.tmpl", gin.H{
			"Run":       run,
			"RunOutput": template.HTML(runOutput),
		})
	}
}

func (web *WebController) ProjectDetailsUpdatePage(c *gin.Context) {
	project, ok := web.validateRequestProjectId(c)
	if !ok {
		return
	}

	message := ""
	projectName := c.PostForm("project_name")
	if projectName != project.Name {
		project.Name = projectName
		models.DB.GormDB.Save(project)
		log.Printf("project name has been updated to %s\n", projectName)
		message = "Project has been updated successfully"
	}

	c.HTML(http.StatusOK, "project_details.tmpl", gin.H{
		"Project": project,
		"Message": message,
	})
}

func (web *WebController) PolicyDetailsUpdatePage(c *gin.Context) {
	policyId64, err := strconv.ParseUint(c.Param("policyid"), 10, 32)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to parse policy id")
		return
	}
	policyId := uint(policyId64)
	policy, ok := models.DB.GetPolicyByPolicyId(c, policyId, middleware.ORGANISATION_ID_KEY)
	if !ok {
		return
	}

	message := ""
	policyText := c.PostForm("policy")
	log.Printf("policyText: %v\n", policyText)
	if policyText != policy.Policy {
		policy.Policy = policyText
		models.DB.GormDB.Save(policy)
		log.Printf("Policy has been updated. policy id: %v\n", policy.ID)
		message = "Policy has been updated successfully"
	}

	c.HTML(http.StatusOK, "policy_details.tmpl", gin.H{
		"Policy":  policy,
		"Message": message,
	})
}

func (web *WebController) RedirectToLoginSubdomain(context *gin.Context) {
	host := context.Request.Host
	hostParts := strings.Split(host, ".")
	if len(hostParts) > 2 {
		hostParts[0] = "login"
		host = strings.Join(hostParts, ".")
	}
	context.Redirect(http.StatusMovedPermanently, fmt.Sprintf("https://%s", host))
}

func (web *WebController) UpdateRepoPage(c *gin.Context) {
	repoId := c.Param("repoid")
	if repoId == "" {
		c.String(http.StatusInternalServerError, "Repo ID can't be empty")
		return
	}
	orgId, exists := c.Get(middleware.ORGANISATION_ID_KEY)
	if !exists {
		log.Printf("Org %v not found in the context\n", middleware.ORGANISATION_ID_KEY)
		c.String(http.StatusInternalServerError, "Not allowed to access this resource")
		return
	}

	repo, err := models.DB.GetRepoById(orgId, repoId)
	if err != nil {
		c.String(http.StatusForbidden, "Failed to find repo")
		return
	}

	if c.Request.Method == "GET" {
		message := ""
		c.HTML(http.StatusOK, "repo_add.tmpl", gin.H{
			"Message": message, "Repo": repo,
		})
		return
	} else if c.Request.Method == "POST" {
		diggerConfigYaml := c.PostForm("diggerconfig")
		if diggerConfigYaml == "" {
			message := "Digger config can't be empty"
			c.HTML(http.StatusOK, "repo_add.tmpl", gin.H{
				"Message": message, "Repo": repo,
			})
			return
		}

		messages, err := models.DB.UpdateRepoDiggerConfig(orgId, diggerConfigYaml, repo)
		if err != nil {
			if strings.HasPrefix(err.Error(), "validation error, ") {
				message := errors.Unwrap(err).Error()
				c.HTML(http.StatusOK, "repo_add.tmpl", gin.H{
					"Message": message, "Repo": repo,
				})
				return
			}
			log.Printf("failed to updated repo %v, %v", repoId, err)
			message := "failed to update repo"
			c.HTML(http.StatusOK, "repo_add.tmpl", gin.H{
				"Message": message, "Repo": repo,
			})
			return
		}
		log.Printf("messages: %v'n", messages)
		c.Redirect(http.StatusFound, "/repos")
	}
}
