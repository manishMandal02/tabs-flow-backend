package spaces

import (
	"net/http"

	"github.com/manishMandal02/tabsflow-backend/pkg/database"
	"github.com/manishMandal02/tabsflow-backend/pkg/http_api"
)

func Router(w http.ResponseWriter, r *http.Request) {

	db := database.New()
	sr := newSpaceRepository(db)
	sh := newSpaceHandler(sr)

	spacesRouter := http_api.NewRouter("/spaces")

	spacesRouter.POST("/", sh.createSpace)
	spacesRouter.GET("/:id", sh.spaceById)
	spacesRouter.GET("/user/:id", sh.spaceById)
	spacesRouter.PATCH("/:id", sh.updateSpace)
	spacesRouter.DELETE("/:id", sh.deleteSpace)

	// serve API routes
	spacesRouter.ServeHTTP(w, r)
}
