package controllers

import (
	"digger.dev/cloud/config"
	"digger.dev/cloud/middleware"
	"digger.dev/cloud/models"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/robert-nix/ansihtml"
	"html/template"
	"net/http"
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
	projects, done := models.GetProjectsFromContext(c, middleware.ORGANISATION_ID_KEY)
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
	projects, done := models.GetProjectsFromContext(c, middleware.ORGANISATION_ID_KEY)
	if !done {
		return
	}

	c.HTML(http.StatusOK, "projects.tmpl", gin.H{
		"Projects": projects,
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
		namespace, ok := models.GetDefaultNamespace(c, middleware.ORGANISATION_ID_KEY)
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

		fmt.Printf("namespace: %v", namespace)
		//TODO: gorm is trying to insert new namespace and organisation on every insert of a new project,
		// there should be a way to avoid it
		project := models.Project{Name: projectName, Organisation: namespace.Organisation, Namespace: namespace}

		err := models.DB.Create(&project).Error
		if err != nil {
			fmt.Printf("Failed to create a new project, %v\n", err)
			message := "Failed to create a project"
			c.HTML(http.StatusOK, "project_add.tmpl", gin.H{
				"Message": message,
			})
		}

		c.Redirect(http.StatusFound, "/projects")
	}
}

func (web *WebController) RunsPage(c *gin.Context) {
	runs, done := models.GetProjectRunsFromContext(c, middleware.ORGANISATION_ID_KEY)
	if !done {
		return
	}
	c.HTML(http.StatusOK, "runs.tmpl", gin.H{
		"Runs": runs,
	})
}

func (web *WebController) PoliciesPage(c *gin.Context) {
	policies, done := models.GetPoliciesFromContext(c, middleware.ORGANISATION_ID_KEY)
	if !done {
		return
	}
	fmt.Println("policies.tmpl")
	c.HTML(http.StatusOK, "policies.tmpl", gin.H{
		"Policies": policies,
	})
}

func (web *WebController) AddPolicyPage(c *gin.Context) {
	if c.Request.Method == "GET" {
		message := ""
		projects, done := models.GetProjectsFromContext(c, middleware.ORGANISATION_ID_KEY)
		if !done {
			return
		}

		policyTypes := make([]string, 0)
		policyTypes = append(policyTypes, "drift")
		policyTypes = append(policyTypes, "terraform")
		policyTypes = append(policyTypes, "access")

		fmt.Printf("projects: %v", projects)

		c.HTML(http.StatusOK, "policy_add.tmpl", gin.H{
			"Message": message, "Projects": projects, "PolicyTypes": policyTypes,
		})
	} else if c.Request.Method == "POST" {
		message := ""
		namespace, ok := models.GetDefaultNamespace(c, middleware.ORGANISATION_ID_KEY)
		if !ok {
			message = "failed to create a new policy"
			c.HTML(http.StatusOK, "policy_add.tmpl", gin.H{
				"Message": message,
			})
		}
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
		project, ok := models.GetProjectByProjectId(c, projectId, middleware.ORGANISATION_ID_KEY)
		if !ok {
			fmt.Printf("Failed to fetch specified project by id: %v, %v\n", projectIdStr, err)
			message := "Failed to create a policy"
			c.HTML(http.StatusOK, "policy_add.tmpl", gin.H{
				"Message": message,
			})
		}

		fmt.Printf("namespace: %v", namespace)

		policy := models.Policy{Project: project, Policy: policyText, Type: policyType, Organisation: namespace.Organisation, Namespace: namespace}

		err = models.DB.Create(&policy).Error
		if err != nil {
			fmt.Printf("Failed to create a new policy, %v\n", err)
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
	policy, ok := models.GetPolicyByPolicyId(c, policyId, middleware.ORGANISATION_ID_KEY)
	if !ok {
		return
	}

	fmt.Println("policy_details.tmpl")
	c.HTML(http.StatusOK, "policy_details.tmpl", gin.H{
		"Policy": policy,
	})
}

func (web *WebController) ProjectDetailsPage(c *gin.Context) {
	project, ok := web.validateRequestProjectId(c)
	if !ok {
		return
	}

	fmt.Println("project_details.tmpl")
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
	run, ok := models.GetProjectByRunId(c, runId, middleware.ORGANISATION_ID_KEY)
	if !ok {
		return
	}

	runOutput := string(ansihtml.ConvertToHTMLWithClasses([]byte(run.Output), "terraform-output-", true))
	runOutput = strings.Replace(runOutput, "  ", "&nbsp;&nbsp;", -1)
	runOutput = strings.Replace(runOutput, "\n", "<br>\n", -1)

	fmt.Println("run_details.tmpl")
	c.HTML(http.StatusOK, "run_details.tmpl", gin.H{
		"Run":       run,
		"RunOutput": template.HTML(runOutput),
	})
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
		models.DB.Save(project)
		fmt.Printf("project name has been updated to %s\n", projectName)
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
	policy, ok := models.GetPolicyByPolicyId(c, policyId, middleware.ORGANISATION_ID_KEY)
	if !ok {
		return
	}

	message := ""
	policyText := c.PostForm("policy")
	fmt.Printf("policyText: %v\n", policyText)
	if policyText != policy.Policy {
		policy.Policy = policyText
		models.DB.Save(policy)
		fmt.Printf("Policy has been updated. policy id: %v", policy.ID)
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
