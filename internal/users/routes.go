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
	usersRouter.DELETE("/:id", handler.deleteUser) // TODO: test delete handler after adding more data

	usersRouter.GET("/:id/preferences", handler.getPreferences)
	usersRouter.PATCH("/:id/preferences", handler.updatePreferences)
	// TODO - subscription
	usersRouter.GET("/:id/subscription", handler.updateUser)
	usersRouter.PATCH("/:id/subscription", handler.updateUser)

	// serve API routes
	usersRouter.ServeHTTP(w, r)
}
