package users

import (
	"net/http"

	"github.com/manishMandal02/tabsflow-backend/pkg/database"
	"github.com/manishMandal02/tabsflow-backend/pkg/http_api"
)

func newMiddleware(ur userRepository) http_api.Handler {
	return func(w http.ResponseWriter, r *http.Request) {
		userId := r.PathValue("id")

		if !checkUserExits(userId, ur, w) {
			return
		}
	}
}

func Router(w http.ResponseWriter, r *http.Request) {

	db := database.New()

	ur := newUserRepository(db)

	handler := newUserHandler(ur)

	usersRouter := http_api.NewRouter("/users")

	checkUserMiddleware := newMiddleware(ur)

	// profile
	usersRouter.POST("/", handler.createUser)
	usersRouter.GET("/:id", handler.userById)
	usersRouter.PATCH("/:id", checkUserMiddleware, handler.updateUser)
	// TODO: test delete handler after adding more data
	usersRouter.DELETE("/:id", checkUserMiddleware, handler.deleteUser)

	// preferences
	usersRouter.GET("/:id/preferences", checkUserMiddleware, handler.getPreferences)
	usersRouter.PATCH("/:id/preferences", checkUserMiddleware, handler.updatePreferences)

	// subscription
	usersRouter.GET("/:id/subscription", checkUserMiddleware, handler.getSubscription)
	usersRouter.GET("/:id/subscription/status", checkUserMiddleware, handler.checkSubscriptionStatus)
	// queries - cancelURL:bool
	usersRouter.GET("/:id/subscription/paddle-url", checkUserMiddleware, handler.getPaddleURL)
	// TODO:test webhook with paddle webhook simulator
	usersRouter.POST("/subscription/webhook", handler.subscriptionWebhook)

	// serve API routes
	usersRouter.ServeHTTP(w, r)
}
