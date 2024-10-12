package notifications

import (
	"net/http"

	"github.com/manishMandal02/tabsflow-backend/pkg/database"
	"github.com/manishMandal02/tabsflow-backend/pkg/http_api"
)

func Router(w http.ResponseWriter, r *http.Request) {
	db := database.New()
	nr := newNotificationRepository(db)
	h := newNotificationHandler(nr)

	// middleware to get userId from jwt token
	userIdMiddleware := newUserIdMiddleware()

	notificationsRouter := http_api.NewRouter("/notifications")

	notificationsRouter.Use(userIdMiddleware)

	notificationsRouter.GET("/:id", h.get)
	notificationsRouter.GET("/my", h.getUserNotifications)
	notificationsRouter.POST("/", h.create)
	notificationsRouter.DELETE("/", h.delete)

	// serve API routes
	notificationsRouter.ServeHTTP(w, r)
}
