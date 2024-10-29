package notifications

import (
	"github.com/manishMandal02/tabsflow-backend/pkg/db"
	"github.com/manishMandal02/tabsflow-backend/pkg/http_api"
)

func Router() http_api.IRouter {
	db := db.New()
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
	notificationsRouter.DELETE("/:id", h.delete)

	// serve API routes
	return notificationsRouter
}
