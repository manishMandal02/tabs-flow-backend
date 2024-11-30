package notes

import (
	"github.com/manishMandal02/tabsflow-backend/pkg/db"
	"github.com/manishMandal02/tabsflow-backend/pkg/events"
	"github.com/manishMandal02/tabsflow-backend/pkg/http_api"
)

func Router(mainTable, searchIndexTable *db.DDB, q *events.Queue) http_api.IRouter {

	nr := NewNoteRepository(mainTable, searchIndexTable)
	nh := newNoteHandler(nr, q)

	// middleware to get userId from jwt token
	userIdMiddleware := newUserIdMiddleware()

	notesRouter := http_api.NewRouter("/notes")

	notesRouter.Use(http_api.SetAllowOriginHeader())

	notesRouter.Use(userIdMiddleware)

	notesRouter.POST("/", nh.create)
	// query: lastNoteId={lastNoteId}
	notesRouter.GET("/my", nh.getAllByUser)
	// query: query={searchTerm}, limit={maxLimit}
	notesRouter.GET("/search", nh.search)

	notesRouter.GET("/:noteId", nh.get)

	notesRouter.PATCH("/", nh.update)

	notesRouter.DELETE("/:noteId", nh.delete)

	// serve API routes
	return notesRouter
}
