package main

import (
	"fmt"
	"gorm.io/gorm"
	"github.com/kmgreen2/agglo/pkg/data"
)

type User struct {
	gorm.Model
	Name string `json:"name"`
	Email string `json:"email"`
}

type IntegrationType string

const (
	JiraIntegration IntegrationType = "jira"
	GithubIntegration = "github"
	CircleCIIntegration = "circleci"
)

type UserIntegration struct {
	gorm.Model
	Type IntegrationType `json:"type" sql:"type:integration_type"`
	UserId int `json:"userId"`
	IntegrationUserId string `json:"integrationUserId"`
}

type TicketState int

const (
	TicketOpened TicketState = iota
	TicketClosed
)

type Ticket struct {
	gorm.Model
	CreatorUserId int `json:"creatorUserId"`
	AssigneeUserId int `json:"assigneeUserId"`
	Title string `json:"title"`
	State TicketState `json:"state" sql:"type:ticket_state"`
}

type Commit struct {
	gorm.Model
	AuthorUserId int `json:"authorUserId"`
	Digest string `json:"digest"`
	Message string `json:"message"`
}

type CICDType string

const (
	CICDBuild CICDType = "build"
	CICDTest = "test"
	CICDRelease = "release"
	CICDDeploy = "deploy"
)

type CICD struct {
	SubmitterUserId int `json:"submitterUserId"`
	CommitDigest string `json:"commitDigest"`
	Type CICDType `json:"type" sql:"type:cicd_type"`
}

type ActivityTrackerModel struct {
	db data.Database
}

func NewActivityTrackerModel(dsn, dbType string) (*ActivityTrackerModel, error) {
	model := &ActivityTrackerModel{}
	config := &data.DatabaseConfig{}

	if dbType != "postgres" {
		return nil, fmt.Errorf("'%s' is not a valid db type", dbType)
	}

	err := config.SetDatabaseType(data.PostgresDatabase)
	if err != nil {
		return nil, err
	}
	err = config.SetDSN(dsn)
	if err != nil {
		return nil, err
	}

	dbImpl, err := data.NewDatabaseImpl(config)
	if err != nil {
		return nil, err
	}

	model.db = dbImpl

	err = model.db.Get().AutoMigrate(&User{}, &UserIntegration{}, &Ticket{}, &Commit{}, &CICD{})
	if err != nil {
		return nil, err
	}

	return model, nil
}

func (model *ActivityTrackerModel) CreateUser(name, email string) error {
	db := model.db.Get()
	return db.Create(&User{
		Name: name,
		Email: email,
	}).Error
}

func (model *ActivityTrackerModel) GetUserByEmail(email string) (*User, error) {
	user := &User{}
	db := model.db.Get()
	if err := db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func (model *ActivityTrackerModel) GetUsers() ([]*User, error) {
	users := make([]*User, 0)
	db := model.db.Get()
	if err := db.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (model *ActivityTrackerModel) CreateUserIntegration(userId int, integrationType IntegrationType,
	integrationUserId string) error {
	db := model.db.Get()
	return db.Create(&UserIntegration {
		Type: integrationType,
		UserId: userId,
		IntegrationUserId: integrationUserId,
	}).Error
}

func (model *ActivityTrackerModel) GetUserFromIntegration(integrationType IntegrationType,
	integrationUserId string) (*User, error) {
	userIntegration := &UserIntegration{}
	user := &User{}
	db := model.db.Get()
	if err := db.Where("integration_user_id = ? and type = ?",
		integrationUserId, string(integrationType)).First(&userIntegration).Error; err != nil {
		return nil, err
	}
	if err := db.Where("id = ?", userIntegration.UserId).First(&user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func (model *ActivityTrackerModel) CreateTicket(creatorUserId, assigneeUserId int, title string,
	state TicketState) error {
	db := model.db.Get()
	return db.Create(&Ticket{
		CreatorUserId: creatorUserId,
		AssigneeUserId: assigneeUserId,
		Title: title,
		State: state,
	}).Error
}

func (model *ActivityTrackerModel) CreateCommit(authorUserId int, digest, message string) error {
	db := model.db.Get()
	return db.Create(&Commit{
		AuthorUserId: authorUserId,
		Digest: digest,
		Message: message,
	}).Error
}

func (model *ActivityTrackerModel) CreateCICD(submitterUserId int, commitDigest string, cicdType CICDType) error {
	db := model.db.Get()
	return db.Create(&CICD{
		SubmitterUserId: submitterUserId,
		CommitDigest: commitDigest,
		Type: cicdType,
	}).Error
}


