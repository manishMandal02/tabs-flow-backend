package users

import (
	"github.com/manishMandal02/tabsflow-backend/pkg/database"
	"github.com/manishMandal02/tabsflow-backend/pkg/http_api"
)

func Router() http_api.IRouter {

	db := database.New()

	ur := newUserRepository(db)

	handler := newUserHandler(ur)

	usersRouter := http_api.NewRouter("/users")

	checkUserMiddleware := newUserMiddleware(ur)

	// preferences
	usersRouter.GET("/preferences", checkUserMiddleware, handler.getPreferences)
	usersRouter.PATCH("/preferences", checkUserMiddleware, handler.updatePreferences)

	// subscription
	usersRouter.GET("/subscription", checkUserMiddleware, handler.getSubscription)
	usersRouter.GET("/subscription/status", checkUserMiddleware, handler.checkSubscriptionStatus)
	// queries - cancelURL:bool
	usersRouter.GET("/subscription/paddle-url", checkUserMiddleware, handler.getPaddleURL)
	usersRouter.POST("/subscription/webhook", handler.subscriptionWebhook)

	// profile
	usersRouter.GET("/:id", handler.userById)
	usersRouter.POST("/", handler.createUser)
	usersRouter.PATCH("/", checkUserMiddleware, handler.updateUser)
	// TODO: test delete handler after adding more data
	usersRouter.DELETE("/", checkUserMiddleware, handler.deleteUser)

	// serve API routes
	return usersRouter
}
