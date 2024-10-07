package spaces

import (
	"net/http"

	"github.com/manishMandal02/tabsflow-backend/pkg/http_api"
)

type spaceHandler struct {
	sr spaceRepository
}

func newSpaceHandler(sr spaceRepository) *spaceHandler {
	return &spaceHandler{
		sr: sr,
	}
}

func (h *spaceHandler) spaceById(w http.ResponseWriter, r *http.Request) {

	http_api.SuccessResMsg(w, "success")
}

func (h *spaceHandler) spacesByUser(w http.ResponseWriter, r *http.Request) {

	http_api.SuccessResMsg(w, "success")
}

func (h *spaceHandler) createSpace(w http.ResponseWriter, r *http.Request) {

	http_api.SuccessResMsg(w, "success")
}

func (h *spaceHandler) updateSpace(w http.ResponseWriter, r *http.Request) {

	http_api.SuccessResMsg(w, "success")
}

func (h *spaceHandler) deleteSpace(w http.ResponseWriter, r *http.Request) {

	http_api.SuccessResMsg(w, "success")
}
