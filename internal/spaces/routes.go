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

	// spaces
	spacesRouter.GET("/:userId/:id", sh.get)
	spacesRouter.POST("/:userId/", sh.create)
	spacesRouter.GET("/:userId", sh.spacesByUser)
	spacesRouter.PATCH("/:userId/:id", sh.update)
	spacesRouter.DELETE("/:id", sh.delete)

	// tabs
	spacesRouter.GET("/tabs/:userId/:spaceId", sh.getTabsInSpace)
	spacesRouter.POST("/tabs/:userId/:spaceId", sh.setTabsInSpace)

	// groups
	spacesRouter.GET("/groups/:userId/:spaceId", sh.getGroupsInSpace)
	spacesRouter.POST("/groups/:userId/:spaceId", sh.setGroupsInSpace)

	// snoozed tabs
	// query param: snoozedAt=unix timestamp
	spacesRouter.GET("/snoozed-tabs/:userId/:spaceId", sh.getSnoozedTabs)
	spacesRouter.POST("/snoozed-tabs/:userId/:spaceId", sh.createSnoozedTab)
	spacesRouter.DELETE("/snoozed-tabs/:userId/:spaceId", sh.deleteSnoozedTab)

	// serve API routes
	spacesRouter.ServeHTTP(w, r)
}
