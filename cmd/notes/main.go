package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/manishMandal02/tabsflow-backend/config"
	"github.com/manishMandal02/tabsflow-backend/internal/notes"
	"github.com/manishMandal02/tabsflow-backend/pkg/db"
	"github.com/manishMandal02/tabsflow-backend/pkg/events"
	"github.com/manishMandal02/tabsflow-backend/pkg/http_api"
)

func main() {

	// load config
	config.Init()

	mainTable := db.New()
	searchIndexTable := db.NewSearchIndexTable()

	queue := events.NewNotificationQueue()

	handler := http_api.NewAPIGatewayHandler("/notes/", notes.Router(mainTable, searchIndexTable, queue))

	lambda.Start(handler.Handle)

}
