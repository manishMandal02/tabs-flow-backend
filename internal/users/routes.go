package users

import (
	"net/http"

	"github.com/manishMandal02/tabsflow-backend/pkg/database"
	"github.com/manishMandal02/tabsflow-backend/pkg/http_api"
)

func Router(w http.ResponseWriter, r *http.Request) {

	db := database.New()

	ur := newUserRepository(db)

	handler := newUserHandler(ur)

	usersRouter := http_api.NewRouter("/users")

	// profile
	usersRouter.POST("/", handler.createUser)
	usersRouter.GET("/:id", handler.userById)
	usersRouter.PATCH("/:id", handler.updateUser)
	// TODO: test delete handler after adding more data
	usersRouter.DELETE("/:id", handler.deleteUser)

	usersRouter.GET("/:id/preferences", handler.getPreferences)
	usersRouter.PATCH("/:id/preferences", handler.updatePreferences)

	usersRouter.GET("/:id/subscription", handler.getSubscription)

	// TODO - implement Paddle webhook with their Go sdk
	usersRouter.POST("/subscription/webhook", handler.updateSubscription)

	// serve API routes
	usersRouter.ServeHTTP(w, r)
}
