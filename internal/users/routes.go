package users

import (
	"github.com/manishMandal02/tabsflow-backend/pkg/db"
	"github.com/manishMandal02/tabsflow-backend/pkg/http_api"
)

func Router() http_api.IRouter {

	db := db.New()

	ur := newUserRepository(db)

	handler := newHandler(ur)

	usersRouter := http_api.NewRouter("/users")

	checkUserMiddleware := newUserMiddleware(ur)

	// profile
	usersRouter.GET("/me", handler.userById)
	usersRouter.POST("/", handler.createUser)
	usersRouter.PATCH("/", checkUserMiddleware, handler.updateUser)
	// TODO: test delete handler after adding more data
	usersRouter.DELETE("/", checkUserMiddleware, handler.deleteUser)

	// preferences
	usersRouter.GET("/preferences", checkUserMiddleware, handler.getPreferences)
	usersRouter.PATCH("/preferences", checkUserMiddleware, handler.updatePreferences)

	// subscription
	usersRouter.GET("/subscription", checkUserMiddleware, handler.getSubscription)
	usersRouter.GET("/subscription/status", checkUserMiddleware, handler.checkSubscriptionStatus)
	// queries - cancelURL:bool
	usersRouter.GET("/subscription/paddle-url", checkUserMiddleware, handler.getPaddleURL)
	usersRouter.POST("/subscription/webhook", handler.subscriptionWebhook)

	// serve API routes
	return usersRouter
}
