package main

import (
	"digger.dev/cloud/models"
	"github.com/dchest/uniuri"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
	"os/exec"
	"testing"
)

func setupSuite(tb testing.TB) (func(tb testing.TB), *models.Database) {
	log.Println("setup suite")

	// database file name
	dbName := "database_test.db"

	// remove old database
	exec.Command("rm", "-f", dbName)

	// open and create a new database
	gdb, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	// migrate tables
	err = gdb.AutoMigrate(&models.Policy{}, &models.Organisation{}, &models.Repo{}, &models.Project{}, &models.Token{},
		&models.User{}, &models.ProjectRun{}, &models.GithubAppInstallation{}, &models.GithubApp{}, &models.GithubAppInstallationLink{},
		&models.GithubDiggerJobLink{}, &models.DiggerJob{})
	if err != nil {
		log.Fatal(err)
	}

	database := &models.Database{GormDB: gdb}

	orgTenantId := "11111111-1111-1111-1111-111111111111"
	externalSource := "test"
	org, err := database.CreateOrganisation(externalSource, orgTenantId)
	if err != nil {
		log.Fatal(err)
	}

	repoName := "test repo"
	repo, err := database.CreateRepo(repoName, org, "")
	if err != nil {
		log.Fatal(err)
	}

	projectName := "test project"
	_, err = database.CreateProject(projectName, org, repo)
	if err != nil {
		log.Fatal(err)
	}

	// Return a function to teardown the test
	return func(tb testing.TB) {
		log.Println("teardown suite")
	}, database
}

func TestCreateDiggerJob(t *testing.T) {
	teardownSuite, database := setupSuite(t)
	defer teardownSuite(t)

	parentJobId := uniuri.New()
	job, err := database.CreateDiggerJob(parentJobId, nil)

	assert.NoError(t, err)
	assert.NotNil(t, job)
	assert.NotZero(t, job.ID)
}

func TestCreateSingleJob(t *testing.T) {
	teardownSuite, database := setupSuite(t)
	defer teardownSuite(t)

	parentJobId := uniuri.New()
	job, err := database.CreateDiggerJob(parentJobId, nil)

	assert.NoError(t, err)
	assert.NotNil(t, job)
	assert.NotZero(t, job.ID)
}
