package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/kmgreen2/agglo/pkg/util"
	"io/ioutil"
	"os"
)

type CommandType int

const (
	CreateUser CommandType = iota
	GetUserByEmail
	GetUsers
	CreateUserIntegration
	GetUserFromIntegration
	CreateTicket
	CreateCommit
	CreateCICD
)

type CommandArgs struct {
	commandType CommandType
	dbType string
	dsn string
}

func usage(msg string, exitCode int) {
	fmt.Println(msg)
	flag.PrintDefaults()
	os.Exit(exitCode)
}

func parseArgs() (*CommandArgs, error){
	var err error
	args := &CommandArgs{}
	commandTypePtr := flag.String("commandType", "", `command types:
	CreateUser CommandType = iota
	GetUserByEmail
	GetUsers
	CreateUserIntegration
	GetUserFromIntegration
	CreateTicket
	CreateCommit
	CreateCICD
	`)
	dsnPtr := flag.String("dsn",
		"host=localhost user=gorm password=gorm dbname=gorm port=5432 sslmode=disable", "Database DSN")
	dbTypePtr := flag.String("dbType", "postgres", "database type (default is postgres")

	flag.Parse()

	args.dbType = *dbTypePtr
	args.dsn = *dsnPtr

	switch *commandTypePtr {
	case "CreateUser":
		args.commandType = CreateUser
	case "GetUserByEmail":
		args.commandType = GetUserByEmail
	case "GetUsers":
		args.commandType = GetUsers
	case "CreateUserIntegration":
		args.commandType = CreateUserIntegration
	case "GetUserFromIntegration":
		args.commandType = GetUserFromIntegration
	case "CreateTicket":
		args.commandType = CreateTicket
	case "CreateCommit":
		args.commandType = CreateCommit
	case "CreateCICD":
		args.commandType = CreateCICD
	default:
		usage(fmt.Sprintf("invalid command type: '%s'", *commandTypePtr), -1)
	}

	return args, err
}

func getStringFromMap(in map[string]interface{}, key string) (string, bool) {
	if entry, ok := in[key]; ok {
		if entryStr, strOk := entry.(string); strOk {
			return entryStr, true
		}
	}
	return "", false
}

func getIntFromMap(in map[string]interface{}, key string) (int, bool) {
	if entry, ok := in[key]; ok {
		if entryFloat, intOk := entry.(float64); intOk {
			return int(entryFloat), true
		}
	}
	return 0, false
}

func main() {
	var inMap, outMap map[string]interface{}

	args, err := parseArgs()
	if err != nil {
		panic(err)
	}

	dbModel, err := NewActivityTrackerModel(args.dsn, args.dbType)
	if err != nil {
		panic(err)
	}

	inBytes, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}

	decodeBuffer := bytes.NewBuffer(inBytes)
	decoder := json.NewDecoder(decodeBuffer)
	err = decoder.Decode(&inMap)
	if err != nil {
		panic(err)
	}

	outMap = util.CopyableMap(inMap).DeepCopy()

	switch args.commandType {
	case CreateUser:
		name, okName := getStringFromMap(inMap, "name")
		if !okName {
			panic("name not specified for user create")
		}
		email, okEmail := getStringFromMap(inMap, "email")
		if !okEmail {
			panic("email not specified for user create")
		}
		err = dbModel.CreateUser(name, email)
		if err != nil {
			panic(err.Error())
		}
	case GetUserByEmail:
		email, okEmail := getStringFromMap(inMap, "email")
		if !okEmail {
			panic("email not specified for user create")
		}
		user, err := dbModel.GetUserByEmail(email)
		if err != nil {
			panic(err.Error())
		}
		outMap["dbUser"] = user
	case GetUsers:
		users, err := dbModel.GetUsers()
		if err != nil {
			panic(err.Error())
		}
		outMap["dbUsers"] = users

	case CreateUserIntegration:
		integrationType, okIntegrationType := getStringFromMap(inMap, "integrationType")
		if !okIntegrationType {
			panic("integrationType not specified for user integration create")
		}

		integrationUserId, okIntegrationUserId := getStringFromMap(inMap, "integrationUserId")
		if !okIntegrationUserId {
			panic("integrationType not specified for user integration create")
		}

		userId, okUserId := getIntFromMap(inMap, "userId")
		if !okUserId {
			panic("userId not specified for user integration create")
		}

		err := dbModel.CreateUserIntegration(userId, IntegrationType(integrationType), integrationUserId)
		if err != nil {
			panic(err.Error())
		}
	case GetUserFromIntegration:
		integrationType, okIntegrationType := getStringFromMap(inMap, "integrationType")
		if !okIntegrationType {
			panic("integrationType not specified for user integration create")
		}

		integrationUserId, okIntegrationUserId := getStringFromMap(inMap, "integrationUserId")
		if !okIntegrationUserId {
			panic("integrationType not specified for user integration create")
		}

		user, err := dbModel.GetUserFromIntegration(IntegrationType(integrationType), integrationUserId)
		if err != nil {
			panic(err.Error())
		}

		outMap["dbUser"] = user
	case CreateTicket:
		integrationType, okIntegrationType := getStringFromMap(inMap, "integrationType")
		if !okIntegrationType {
			panic("integrationType not specified for creating ticket")
		}

		creatorId, okCreatorId := getStringFromMap(inMap, "creator")
		if !okCreatorId {
			panic("creator not specified when creating ticket")
		}

		creatorUser, err := dbModel.GetUserFromIntegration(IntegrationType(integrationType), creatorId)
		if err != nil {
			panic(err.Error())
		}

		assigneeId, okAssigneeId := getStringFromMap(inMap, "assignee")
		if !okAssigneeId {
			panic("assignee not specified when creating ticket")
		}

		assigneeUser, err := dbModel.GetUserFromIntegration(IntegrationType(integrationType), assigneeId)
		if err != nil {
			panic(err.Error())
		}

		title, okTitle := getStringFromMap(inMap, "title")
		if !okTitle {
			panic("title not specified when creating ticket")
		}

		state, okState := getStringFromMap(inMap, "state")
		if !okState {
			panic("state not specified when creating ticket")
		}

		status := TicketOpened
		if state == "Done" {
			status = TicketClosed
		}

		err = dbModel.CreateTicket(int(creatorUser.ID), int(assigneeUser.ID), title, status)
		if err != nil {
			panic(err.Error())
		}

	case CreateCommit:
		integrationType, okIntegrationType := getStringFromMap(inMap, "integrationType")
		if !okIntegrationType {
			panic("integrationType not specified for creating ticket")
		}

		author, okAuthor := getStringFromMap(inMap, "author")
		if !okAuthor {
			panic("author not specified when creating ticket")
		}

		authorUser, err := dbModel.GetUserFromIntegration(IntegrationType(integrationType), author)
		if err != nil {
			panic(err.Error())
		}

		digest, okDigest := getStringFromMap(inMap, "digest")
		if !okDigest {
			panic("digest not specified when creating ticket")
		}

		message, okMessage := getStringFromMap(inMap, "message")
		if !okMessage {
			panic("message not specified when creating ticket")
		}

		err = dbModel.CreateCommit(int(authorUser.ID), digest, message)
		if err != nil {
			panic(err.Error())
		}

	case CreateCICD:
		integrationType, okIntegrationType := getStringFromMap(inMap, "integrationType")
		if !okIntegrationType {
			panic("integrationType not specified for creating CICD event")
		}

		submitter, okSubmitter := getStringFromMap(inMap, "submitter")
		if !okSubmitter {
			panic("submitter not specified when creating CICD event")
		}

		submitterUser, err := dbModel.GetUserFromIntegration(IntegrationType(integrationType), submitter)
		if err != nil {
			panic(err.Error())
		}

		digest, okDigest := getStringFromMap(inMap, "digest")
		if !okDigest {
			panic("digest not specified when creating ticket")
		}

		eventType, okEventType := getStringFromMap(inMap, "eventType")
		if !okEventType {
			panic("eventType not specified when creating ticket")
		}

		err = dbModel.CreateCICD(int(submitterUser.ID), digest, CICDType(eventType))
		if err != nil {
			panic(err.Error())
		}
	}

	encodeBuffer := bytes.NewBuffer([]byte{})
	encoder := json.NewEncoder(encodeBuffer)
	err = encoder.Encode(outMap)
	if err != nil {
		panic(err)
	}
	fmt.Print(encodeBuffer.String())
}