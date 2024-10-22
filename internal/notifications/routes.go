package notifications

import (
	"net/http"

	"github.com/manishMandal02/tabsflow-backend/pkg/database"
	"github.com/manishMandal02/tabsflow-backend/pkg/http_api"
)

func Router(w http.ResponseWriter, r *http.Request) {
	db := database.New()
	nr := newRepository(db)
	h := newHandler(nr)

	// middleware to get userId from jwt token
	userIdMiddleware := newUserIdMiddleware()

	notificationsRouter := http_api.NewRouter("/notifications")

	notificationsRouter.Use(userIdMiddleware)

	// notifications subscription
	notificationsRouter.GET("/subscription", h.getNotificationSubscription)
	notificationsRouter.POST("/subscription", h.subscribe)
	notificationsRouter.DELETE("/subscription", h.unsubscribe)

	notificationsRouter.GET("/my", h.getUserNotifications)
	notificationsRouter.GET("/:id", h.get)
	notificationsRouter.DELETE("/", h.delete)

	// serve API routes
	notificationsRouter.ServeHTTP(w, r)
}
