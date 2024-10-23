package notes

import (
	"github.com/manishMandal02/tabsflow-backend/pkg/database"
	"github.com/manishMandal02/tabsflow-backend/pkg/http_api"
)

func Router() http_api.IRouter {
	db := database.New()
	searchIndexTable := database.NewSearchIndexTable()
	nr := NewNoteRepository(db, searchIndexTable)
	nh := newNoteHandler(nr)

	// middleware to get userId from jwt token
	userIdMiddleware := newUserIdMiddleware()

	notesRouter := http_api.NewRouter("/notes")

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
