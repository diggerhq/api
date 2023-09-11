package main

import (
	"digger.dev/cloud/models"
	"github.com/dchest/uniuri"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
	"os"
	"testing"
)

func setupSuite(tb testing.TB) (func(tb testing.TB), *models.Database) {
	log.Println("setup suite")

	// database file name
	dbName := "database_test.db"

	// remove old database
	e := os.Remove(dbName)
	if e != nil {
		log.Fatal(e)
	}

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
	orgName := "testOrg"
	org, err := database.CreateOrganisation(orgName, externalSource, orgTenantId)
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

func TestFindDiggerJobsByParentJobId(t *testing.T) {
	teardownSuite, database := setupSuite(t)
	defer teardownSuite(t)

	parentJobId := uniuri.New()
	job1Id := uniuri.New()
	job2Id := uniuri.New()
	job, err := database.CreateDiggerJob(parentJobId, nil)
	assert.NoError(t, err)
	assert.NotNil(t, job)
	assert.NotZero(t, job.ID)
	job, err = database.CreateDiggerJob(job1Id, &parentJobId)
	assert.NoError(t, err)
	assert.NotNil(t, job)
	assert.Equal(t, parentJobId, *job.ParentDiggerJobId)
	assert.NotZero(t, job.ID)
	job, err = database.CreateDiggerJob(job2Id, &parentJobId)
	assert.NoError(t, err)
	assert.NotNil(t, job)
	assert.Equal(t, parentJobId, *job.ParentDiggerJobId)
	assert.NotZero(t, job.ID)

	jobs, err := database.GetDiggerJobsByParentId(parentJobId)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(jobs))
	assert.Equal(t, job1Id, jobs[0].DiggerJobId)
	assert.Equal(t, job2Id, jobs[1].DiggerJobId)
}
