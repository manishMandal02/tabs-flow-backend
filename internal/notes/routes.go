package notes

import (
	"net/http"

	"github.com/manishMandal02/tabsflow-backend/pkg/database"
	"github.com/manishMandal02/tabsflow-backend/pkg/http_api"
)

func Router(w http.ResponseWriter, r *http.Request) {
	db := database.New()
	nr := newNoteRepository(db)
	nh := newNoteHandler(nr)

	notesRouter := http_api.NewRouter("/notes")

	notesRouter.GET("/:userId", nh.getNotes)
	notesRouter.POST("/:userId", nh.createNote)
	notesRouter.PATCH("/:userId", nh.updateNote)

	notesRouter.DELETE("/:userId/:noteId", nh.deleteNote)

	// serve API routes
	notesRouter.ServeHTTP(w, r)
}
