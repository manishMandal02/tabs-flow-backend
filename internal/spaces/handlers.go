package spaces

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/manishMandal02/tabsflow-backend/pkg/http_api"
	"github.com/manishMandal02/tabsflow-backend/pkg/logger"
)

type spaceHandler struct {
	sr spaceRepository
}

func newSpaceHandler(sr spaceRepository) *spaceHandler {
	return &spaceHandler{
		sr: sr,
	}
}

func (h *spaceHandler) get(w http.ResponseWriter, r *http.Request) {
	userId := r.PathValue("userId")
	spaceId := r.PathValue("id")

	if userId == "" || spaceId == "" {
		http.Error(w, errMsg.groupsSet, http.StatusBadRequest)
		return
	}

	space, err := h.sr.getSpaceById(userId, spaceId)

	if err != nil {
		logger.Error("error getting space", err)
		http.Error(w, errMsg.spaceGet, http.StatusBadGateway)
		return
	}

	http_api.SuccessResData(w, space)
}

func (h *spaceHandler) spacesByUser(w http.ResponseWriter, r *http.Request) {
	userId := r.PathValue("userId")

	if userId == "" {
		http.Error(w, errMsg.groupsSet, http.StatusBadRequest)
		return
	}

	spaces, err := h.sr.getSpacesByUser(userId)

	if err != nil {
		logger.Error("error getting spaces", err)
		http.Error(w, errMsg.spaceGet, http.StatusBadGateway)
		return
	}

	http_api.SuccessResData(w, spaces)

	http_api.SuccessResMsg(w, "success")
}

func (h *spaceHandler) create(w http.ResponseWriter, r *http.Request) {

	userId := r.PathValue("userId")

	s := space{}

	err := json.NewDecoder(r.Body).Decode(&s)

	if err != nil {
		http.Error(w, errMsg.spaceCreate, http.StatusBadRequest)
		return
	}

	err = s.validate()

	if err != nil {
		logger.Error("error validating space", err)
		http.Error(w, errMsg.spaceCreate, http.StatusBadRequest)
		return
	}

	err = h.sr.createSpace(userId, &s)

	if err != nil {
		logger.Error("error creating space", err)
		http.Error(w, errMsg.spaceCreate, http.StatusBadRequest)
		return
	}

	http_api.SuccessResMsg(w, "space created successfully")
}

func (h *spaceHandler) update(w http.ResponseWriter, r *http.Request) {

	userId := r.PathValue("userId")
	spaceId := r.PathValue("id")

	if userId == "" || spaceId == "" {
		http.Error(w, errMsg.groupsSet, http.StatusBadRequest)
		return
	}

	s := space{}

	err := json.NewDecoder(r.Body).Decode(&s)

	if err != nil {
		logger.Error("error decoding space", err)
		http.Error(w, errMsg.spaceUpdate, http.StatusBadRequest)
		return
	}

	err = h.sr.updateSpace(userId, spaceId, &s)

	if err != nil {
		logger.Error("error updating space", err)
		http.Error(w, errMsg.spaceUpdate, http.StatusBadGateway)
		return
	}

	http_api.SuccessResMsg(w, "space updated successfully")
}

func (h *spaceHandler) delete(w http.ResponseWriter, r *http.Request) {

	spaceId := r.PathValue("id")
	userId := r.PathValue("userId")

	if userId == "" || spaceId == "" {
		http.Error(w, errMsg.groupsSet, http.StatusBadRequest)
		return
	}

	err := h.sr.deleteSpace(userId, spaceId)

	if err != nil {
		logger.Error("error deleting space", err)
		http.Error(w, errMsg.spaceDelete, http.StatusBadGateway)
		return
	}
	http_api.SuccessResMsg(w, "space deleted successfully")
}

func (h *spaceHandler) getTabsInSpace(w http.ResponseWriter, r *http.Request) {
	userId := r.PathValue("userId")
	spaceId := r.PathValue("spaceId")

	tabs, err := h.sr.getTabsForSpace(userId, spaceId)

	if err != nil {
		logger.Error("error getting tabs for space", err)
		http.Error(w, errMsg.tabsGet, http.StatusBadGateway)
		return
	}

	http_api.SuccessResData(w, tabs)
}

func (h *spaceHandler) setTabsInSpace(w http.ResponseWriter, r *http.Request) {
	userId := r.PathValue("userId")
	spaceId := r.PathValue("spaceId")

	if userId == "" || spaceId == "" {
		http.Error(w, errMsg.groupsSet, http.StatusBadRequest)
		return
	}

	tabs := []tab{}

	err := json.NewDecoder(r.Body).Decode(&tabs)

	if err != nil {
		logger.Error("error decoding tabs", err)
		http.Error(w, errMsg.tabsSet, http.StatusBadRequest)
		return
	}

	err = h.sr.setTabsForSpace(userId, spaceId, &tabs)

	if err != nil {
		logger.Error("error setting tabs for space", err)
		http.Error(w, errMsg.tabsSet, http.StatusBadGateway)
		return
	}

	http_api.SuccessResMsg(w, "tabs set successfully")
}

func (h *spaceHandler) getGroupsInSpace(w http.ResponseWriter, r *http.Request) {
	userId := r.PathValue("userId")
	spaceId := r.PathValue("spaceId")

	if userId == "" || spaceId == "" {
		http.Error(w, errMsg.groupsSet, http.StatusBadRequest)
		return
	}

	groups, err := h.sr.getGroupsForSpace(userId, spaceId)

	if err != nil {
		logger.Error("error getting groups for space", err)
		http.Error(w, errMsg.groupsGet, http.StatusBadGateway)
		return
	}

	http_api.SuccessResData(w, groups)
}

func (h *spaceHandler) setGroupsInSpace(w http.ResponseWriter, r *http.Request) {
	userId := r.PathValue("userId")
	spaceId := r.PathValue("spaceId")

	if userId == "" || spaceId == "" {
		http.Error(w, errMsg.groupsSet, http.StatusBadRequest)
		return
	}

	groups := []group{}

	err := json.NewDecoder(r.Body).Decode(&groups)

	if err != nil {
		logger.Error("error decoding groups", err)
		http.Error(w, errMsg.groupsSet, http.StatusBadRequest)
		return
	}

	err = h.sr.setGroupsForSpace(userId, spaceId, &groups)

	if err != nil {
		logger.Error("error setting groups for space", err)
		http.Error(w, errMsg.groupsSet, http.StatusBadGateway)
		return
	}

	http_api.SuccessResMsg(w, "groups set successfully")

}

func (h *spaceHandler) createSnoozedTab(w http.ResponseWriter, r *http.Request) {
	userId := r.PathValue("userId")
	spaceId := r.PathValue("spaceId")

	sT := snoozedTab{}

	err := json.NewDecoder(r.Body).Decode(&sT)

	if err != nil {
		logger.Error("error decoding snoozed tab", err)
		http.Error(w, errMsg.snoozedTabsCreate, http.StatusBadRequest)
		return
	}

	err = h.sr.addSnoozedTab(userId, spaceId, &sT)

	if err != nil {
		logger.Error("error snoozing tab", err)
		http.Error(w, errMsg.snoozedTabsCreate, http.StatusBadGateway)
		return
	}

	http_api.SuccessResMsg(w, "tab snoozed successfully")
}

func (h *spaceHandler) getSnoozedTabs(w http.ResponseWriter, r *http.Request) {
	userId := r.PathValue("userId")
	spaceId := r.PathValue("spaceId")

	if userId == "" || spaceId == "" {
		http.Error(w, errMsg.snoozedTabsGet, http.StatusBadRequest)
		return
	}

	snoozedAt := r.URL.Query().Get("snoozedAt")

	if snoozedAt == "" {

		lastKey := r.URL.Query().Get("lastSnoozedTabId")

		if lastKey == "" {
			lastKey = "0"
		}

		lastSnoozedTabId, err := strconv.ParseInt(lastKey, 10, 64)

		if err != nil {
			logger.Error("error parsing lastSnoozedTabId", err)
			http.Error(w, errMsg.snoozedTabsGet, http.StatusBadRequest)
			return
		}

		// return all snoozed tabs for space
		sT, err := h.sr.geSnoozedTabsInSpace(userId, spaceId, lastSnoozedTabId)

		if err != nil {
			logger.Error("error getting snoozed tabs for space", err)
			http.Error(w, errMsg.snoozedTabsGet, http.StatusBadGateway)
			return
		}
		http_api.SuccessResData(w, sT)
		return
	}

	// return snoozed tab that matches snoozedAt time

	snoozedAtInt, err := strconv.ParseInt(snoozedAt, 10, 64)

	if err != nil {
		logger.Error("error parsing snoozedAt", err)
		http.Error(w, errMsg.snoozedTabsGet, http.StatusBadRequest)
		return
	}

	sT, err := h.sr.getSnoozedTab(userId, spaceId, snoozedAtInt)

	if err != nil {
		logger.Error("error getting snoozed tab", err)
		http.Error(w, errMsg.snoozedTabsGet, http.StatusBadGateway)
		return
	}

	http_api.SuccessResData(w, sT)
}

func (h *spaceHandler) deleteSnoozedTab(w http.ResponseWriter, r *http.Request) {
	userId := r.PathValue("userId")
	spaceId := r.PathValue("spaceId")

	snoozedAt := r.URL.Query().Get("snoozedAt")

	if userId == "" || spaceId == "" || snoozedAt == "" {
		http.Error(w, errMsg.snoozedTabsDelete, http.StatusBadRequest)
		return
	}

	snoozedAtInt, err := strconv.ParseInt(snoozedAt, 10, 64)

	if err != nil {
		logger.Error("error parsing snoozedAt", err)
		http.Error(w, errMsg.snoozedTabsDelete, http.StatusBadRequest)
		return
	}

	err = h.sr.deleteSnoozedTab(userId, spaceId, snoozedAtInt)

	if err != nil {
		logger.Error("error deleting snoozed tab", err)
		http.Error(w, errMsg.snoozedTabsDelete, http.StatusBadGateway)
		return
	}

	http_api.SuccessResMsg(w, "snoozed tab deleted successfully")
}
