package notes

import (
	"net/http"

	"github.com/manishMandal02/tabsflow-backend/pkg/database"
	"github.com/manishMandal02/tabsflow-backend/pkg/http_api"
)

func Router(w http.ResponseWriter, r *http.Request) {
	db := database.New()
	searchIndexTable := database.NewSearchIndexTable()
	nr := newNoteRepository(db, searchIndexTable)
	nh := newNoteHandler(nr)

	notesRouter := http_api.NewRouter("/notes")

	notesRouter.GET("/:userId/:noteId", nh.get)
	// query: lastNoteId={lastNoteId}
	notesRouter.GET("/:userId", nh.getAllByUser)
	// query: query={searchTerm}, limit={maxLimit}
	notesRouter.GET("/:userId/search", nh.search)
	notesRouter.POST("/:userId", nh.create)
	notesRouter.PATCH("/:userId", nh.update)

	notesRouter.DELETE("/:userId/:noteId", nh.delete)

	// serve API routes
	notesRouter.ServeHTTP(w, r)
}
