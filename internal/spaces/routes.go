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

	// middleware to get userId from jwt token
	userIdMiddleware := newUserIdMiddleware()

	spacesRouter := http_api.NewRouter("/spaces")

	spacesRouter.Use(userIdMiddleware)

	// spaces
	spacesRouter.POST("/", sh.create)
	spacesRouter.GET("/:id", sh.get)
	spacesRouter.GET("/my", sh.spacesByUser)
	spacesRouter.PATCH("/", sh.update)
	spacesRouter.DELETE("/:spaceId", sh.delete)

	// tabs
	spacesRouter.GET("/:spaceId/tabs", sh.getTabsInSpace)
	spacesRouter.POST("/:spaceId/tabs", sh.setTabsInSpace)

	// groups
	spacesRouter.GET("/:spaceId/groups", sh.getGroupsInSpace)
	spacesRouter.POST("/:spaceId/groups", sh.setGroupsInSpace)

	// snoozed tabs
	spacesRouter.POST("/:spaceId/snoozed-tabs", sh.createSnoozedTab)
	// query param: snoozedAt={timestamp}
	spacesRouter.GET("/:spaceId/snoozed-tabs", sh.getSnoozedTabs)
	spacesRouter.DELETE("/:spaceId/snoozed-tabs", sh.deleteSnoozedTab)

	// serve API routes
	spacesRouter.ServeHTTP(w, r)
}
