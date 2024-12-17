package spaces

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/manishMandal02/tabsflow-backend/pkg/events"
	"github.com/manishMandal02/tabsflow-backend/pkg/http_api"
	"github.com/manishMandal02/tabsflow-backend/pkg/logger"
)

type spaceHandler struct {
	r                 spaceRepository
	notificationQueue *events.Queue
}

func newSpaceHandler(r spaceRepository, q *events.Queue) *spaceHandler {
	return &spaceHandler{
		r:                 r,
		notificationQueue: q,
	}
}

func (h *spaceHandler) setUserDefaultSpace(w http.ResponseWriter, r *http.Request) {
	userId := r.Header.Get("UserId")

	err := h.r.createSpace(userId, &defaultUserSpace)

	if err != nil {
		http_api.ErrorRes(w, errMsg.userDefaultSpace, http.StatusBadGateway)
		return
	}

	m := &http_api.Metadata{
		UpdatedAt: time.Now().UnixMilli(),
	}

	err = h.r.setGroupsForSpace(userId, defaultSpaceId, &defaultUserGroup, m)

	if err != nil {
		http_api.ErrorRes(w, errMsg.userDefaultSpace, http.StatusBadGateway)
		return
	}

	err = h.r.setTabsForSpace(userId, defaultSpaceId, &defaultUserTabs, m)

	if err != nil {
		http_api.ErrorRes(w, errMsg.userDefaultSpace, http.StatusBadGateway)
		return
	}

	http_api.SuccessResMsg(w, "user default space created successfully")

}

func (h *spaceHandler) get(w http.ResponseWriter, r *http.Request) {
	userId := r.PathValue("userId")
	spaceId := r.PathValue("id")

	if spaceId == "" {
		http_api.ErrorRes(w, errMsg.spaceId, http.StatusBadRequest)
		return
	}

	space, err := h.r.getSpaceById(userId, spaceId)

	if err != nil {

		if err.Error() == errMsg.spaceNotFound {
			//  space not found
			http_api.ErrorRes(w, errMsg.spaceNotFound, http.StatusNotFound)
			return
		}
		http_api.ErrorRes(w, errMsg.spaceGet, http.StatusBadGateway)
		return
	}

	http_api.SuccessResData(w, space)
}

func (h *spaceHandler) spacesByUser(w http.ResponseWriter, r *http.Request) {
	userId := r.PathValue("userId")

	if userId == "" {
		http_api.ErrorRes(w, errMsg.groupsSet, http.StatusBadRequest)
		return
	}

	spaces, err := h.r.getSpacesByUser(userId)

	if err != nil {
		if err.Error() == errMsg.spaceNotFound {
			http_api.SuccessResData(w, []string{})
			return
		}
		logger.Error("error getting spaces", err)
		http_api.ErrorRes(w, errMsg.spaceGet, http.StatusBadGateway)
		return
	}

	http_api.SuccessResData(w, spaces)

}

func (h *spaceHandler) create(w http.ResponseWriter, r *http.Request) {

	userId := r.PathValue("userId")

	s := space{}

	err := json.NewDecoder(r.Body).Decode(&s)

	if err != nil {
		logger.Error("error un_marshalling body", err)
		http_api.ErrorRes(w, errMsg.spaceCreate, http.StatusBadRequest)
		return
	}

	// validate space
	err = s.validate()

	if err != nil {
		logger.Error("error validating space", err)
		http_api.ErrorRes(w, errMsg.spaceCreate, http.StatusBadRequest)
		return
	}

	err = h.r.createSpace(userId, &s)

	if err != nil {
		logger.Error("error creating space", err)
		http_api.ErrorRes(w, errMsg.spaceCreate, http.StatusBadRequest)
		return
	}

	http_api.SuccessResMsg(w, "space created successfully")
}

func (h *spaceHandler) update(w http.ResponseWriter, r *http.Request) {

	userId := r.PathValue("userId")

	s := space{}

	err := json.NewDecoder(r.Body).Decode(&s)

	if err != nil {
		logger.Error("error decoding space", err)
		http_api.ErrorRes(w, errMsg.spaceUpdate, http.StatusBadRequest)
		return
	}

	if s.Id == "" {
		http_api.ErrorRes(w, errMsg.spaceId, http.StatusBadRequest)
		return
	}

	_, err = h.r.getSpaceById(userId, s.Id)

	if err != nil {
		if err.Error() == errMsg.spaceNotFound {
			//  space not found
			http_api.ErrorRes(w, errMsg.spaceNotFound, http.StatusBadRequest)
			return
		}
		http_api.ErrorRes(w, errMsg.spaceUpdate, http.StatusBadGateway)
		return
	}

	err = h.r.updateSpace(userId, &s)

	if err != nil {
		logger.Error("error updating space", err)
		http_api.ErrorRes(w, errMsg.spaceUpdate, http.StatusBadGateway)
		return
	}

	http_api.SuccessResMsg(w, "space updated successfully")
}

func (h *spaceHandler) delete(w http.ResponseWriter, r *http.Request) {

	spaceId := r.PathValue("spaceId")
	userId := r.PathValue("userId")

	if spaceId == "" {
		http_api.ErrorRes(w, errMsg.spaceId, http.StatusBadRequest)
		return
	}

	err := h.r.deleteSpace(userId, spaceId)

	if err != nil {
		logger.Error("error deleting space", err)
		http_api.ErrorRes(w, errMsg.spaceDelete, http.StatusBadGateway)
		return
	}
	http_api.SuccessResMsg(w, "space deleted successfully")
}

// tabs
func (h *spaceHandler) getTabsInSpace(w http.ResponseWriter, r *http.Request) {
	userId := r.PathValue("userId")
	spaceId := r.PathValue("spaceId")

	tabs, m, err := h.r.getTabsForSpace(userId, spaceId)

	if err != nil {
		logger.Error("error getting tabs for space", err)
		http_api.ErrorRes(w, errMsg.tabsGet, http.StatusBadGateway)
		return
	}

	http_api.SuccessResDataWithMetadata(w, tabs, m)
}

func (h *spaceHandler) setTabsInSpace(w http.ResponseWriter, r *http.Request) {
	userId := r.PathValue("userId")
	spaceId := r.PathValue("spaceId")

	if spaceId == "" {
		http_api.ErrorRes(w, errMsg.spaceId, http.StatusBadRequest)
		return
	}

	data := struct {
		Tabs []tab `json:"tabs"`
	}{}

	err := json.NewDecoder(r.Body).Decode(&data)

	if err != nil {
		logger.Error("error decoding tabs", err)
		http_api.ErrorRes(w, errMsg.tabsSet, http.StatusBadRequest)
		return
	}

	if len(data.Tabs) < 1 {
		http_api.ErrorRes(w, errMsg.tabsSet, http.StatusBadRequest)
		return
	}

	m := &http_api.Metadata{
		UpdatedAt: time.Now().UnixMilli(),
	}

	err = h.r.setTabsForSpace(userId, spaceId, &data.Tabs, m)

	if err != nil {
		logger.Error("error setting tabs for space", err)
		http_api.ErrorRes(w, errMsg.tabsSet, http.StatusBadGateway)
		return
	}

	http_api.SuccessResMsgWithMetadata(w, "tabs set successfully", m)
}

// groups
func (h *spaceHandler) getGroupsInSpace(w http.ResponseWriter, r *http.Request) {
	userId := r.PathValue("userId")
	spaceId := r.PathValue("spaceId")

	if spaceId == "" {
		http_api.ErrorRes(w, errMsg.spaceId, http.StatusBadRequest)
		return
	}

	groups, m, err := h.r.getGroupsForSpace(userId, spaceId)

	if err != nil {
		logger.Error("error getting groups for space", err)
		http_api.ErrorRes(w, errMsg.groupsGet, http.StatusBadGateway)
		return
	}

	http_api.SuccessResDataWithMetadata(w, groups, m)
}

func (h *spaceHandler) setGroupsInSpace(w http.ResponseWriter, r *http.Request) {
	userId := r.PathValue("userId")
	spaceId := r.PathValue("spaceId")

	if spaceId == "" {
		http_api.ErrorRes(w, errMsg.spaceId, http.StatusBadRequest)
		return
	}
	data := struct {
		Groups []group `json:"groups"`
	}{}

	err := json.NewDecoder(r.Body).Decode(&data)

	if err != nil {
		logger.Error("error decoding groups", err)
		http_api.ErrorRes(w, errMsg.groupsSet, http.StatusBadRequest)
		return
	}

	if len(data.Groups) < 1 {
		http_api.ErrorRes(w, errMsg.groupsSet, http.StatusBadRequest)
		return
	}

	m := &http_api.Metadata{
		UpdatedAt: time.Now().UnixMilli(),
	}

	err = h.r.setGroupsForSpace(userId, spaceId, &data.Groups, m)

	if err != nil {
		logger.Error("error setting groups for space", err)
		http_api.ErrorRes(w, errMsg.groupsSet, http.StatusBadGateway)
		return
	}

	http_api.SuccessResMsgWithMetadata(w, "groups set successfully", m)

}

// snoozed tabs
func (h *spaceHandler) createSnoozedTab(w http.ResponseWriter, r *http.Request) {
	userId := r.PathValue("userId")
	spaceId := r.PathValue("spaceId")

	sT := SnoozedTab{}

	err := json.NewDecoder(r.Body).Decode(&sT)

	if err != nil {
		logger.Error("error decoding snoozed tab", err)
		http_api.ErrorRes(w, errMsg.snoozedTabsCreate, http.StatusBadRequest)
		return
	}

	err = h.r.addSnoozedTab(userId, spaceId, &sT)

	if err != nil {
		logger.Error("error snoozing tab", err)
		http_api.ErrorRes(w, errMsg.snoozedTabsCreate, http.StatusBadGateway)
		return
	}

	// create a schedule for the tab, to un-snooze the tab
	event := events.New(events.EventTypeScheduleSnoozedTab, &events.ScheduleSnoozedTabPayload{
		UserId:       userId,
		SpaceId:      spaceId,
		SnoozedTabId: strconv.FormatInt(sT.SnoozedAt, 10),
		SubEvent:     events.SubEventCreate,
		TriggerAt:    sT.SnoozedUntil,
	})

	err = h.notificationQueue.AddMessage(event)

	if err != nil {
		http_api.ErrorRes(w, errMsg.snoozedTabsCreate, http.StatusBadGateway)
		return
	}

	http_api.SuccessResMsg(w, "tab snoozed successfully")
}

func (h *spaceHandler) getSnoozedTab(w http.ResponseWriter, r *http.Request) {
	userId := r.PathValue("userId")
	spaceId := r.PathValue("spaceId")
	snoozedTabId := r.PathValue("id")

	if spaceId == "" || snoozedTabId == "" {
		http_api.ErrorRes(w, errMsg.spaceId, http.StatusBadRequest)
		return
	}

	intId, err := strconv.ParseInt(snoozedTabId, 10, 64)

	if err != nil {
		logger.Error("error parsing snoozedTabId to int", err)
		http_api.ErrorRes(w, errMsg.snoozedTabsGet, http.StatusBadGateway)
		return
	}

	sT, err := h.r.GetSnoozedTab(userId, spaceId, intId)

	if err != nil {
		if err.Error() == errMsg.snoozedTabsNotFound {
			http_api.ErrorRes(w, errMsg.snoozedTabsNotFound, http.StatusNotFound)
			return
		}
		logger.Error("error getting snoozed tab", err)
		http_api.ErrorRes(w, errMsg.snoozedTabsGet, http.StatusBadGateway)
		return
	}

	http_api.SuccessResData(w, sT)
}

func (h *spaceHandler) getSnoozedTabsBySpace(w http.ResponseWriter, r *http.Request) {
	userId := r.PathValue("userId")
	spaceId := r.PathValue("spaceId")

	if spaceId == "" {
		http_api.ErrorRes(w, errMsg.spaceId, http.StatusBadRequest)
		return
	}

	lastKey := r.URL.Query().Get("lastSnoozedTabId")

	if lastKey == "" {
		lastKey = "0"
	}

	lastSnoozedTabId, err := strconv.ParseInt(lastKey, 10, 64)

	if err != nil {
		logger.Error("error parsing lastSnoozedTabId", err)
		http_api.ErrorRes(w, errMsg.snoozedTabsGet, http.StatusBadRequest)
		return
	}

	// return all snoozed tabs for space
	sT, err := h.r.geSnoozedTabsInSpace(userId, spaceId, lastSnoozedTabId)

	if err != nil {
		logger.Error("error getting snoozed tabs for space", err)
		http_api.ErrorRes(w, errMsg.snoozedTabsGet, http.StatusBadGateway)
		return
	}
	http_api.SuccessResData(w, sT)
}

func (h spaceHandler) getSnoozedTabByUser(w http.ResponseWriter, r *http.Request) {
	userId := r.PathValue("userId")

	lastKey := r.URL.Query().Get("lastSnoozedTabId")

	if lastKey == "" {
		lastKey = "0"
	}

	lastSnoozedTabId, err := strconv.ParseInt(lastKey, 10, 64)

	if err != nil {
		logger.Error("error parsing lastSnoozedTabId", err)
		http_api.ErrorRes(w, errMsg.snoozedTabsGet, http.StatusBadRequest)
		return
	}

	sT, err := h.r.getAllSnoozedTabsByUser(userId, lastSnoozedTabId)

	if err != nil {
		if err.Error() == errMsg.snoozedTabsNotFound {
			http_api.SuccessResData(w, []SnoozedTab{})

			return
		}
		logger.Error("error getting snoozed tabs for user", err)
		http_api.ErrorRes(w, errMsg.snoozedTabsGet, http.StatusBadGateway)
		return
	}

	http_api.SuccessResData(w, sT)
}

func (h *spaceHandler) DeleteSnoozedTab(w http.ResponseWriter, r *http.Request) {
	userId := r.PathValue("userId")
	spaceId := r.PathValue("spaceId")
	snoozedAt := r.PathValue("id")

	if spaceId == "" || snoozedAt == "" {
		http_api.ErrorRes(w, errMsg.spaceId, http.StatusBadRequest)
		return
	}

	snoozedAtInt, err := strconv.ParseInt(snoozedAt, 10, 64)

	if err != nil {
		logger.Error("error parsing snoozedAt", err)
		http_api.ErrorRes(w, errMsg.snoozedTabsDelete, http.StatusBadRequest)
		return
	}

	err = h.r.DeleteSnoozedTab(userId, spaceId, snoozedAtInt)

	if err != nil {
		logger.Error("error deleting snoozed tab", err)
		http_api.ErrorRes(w, errMsg.snoozedTabsDelete, http.StatusBadGateway)
		return
	}

	//  delete notification the schedule
	event := events.New(events.EventTypeScheduleSnoozedTab, &events.ScheduleSnoozedTabPayload{
		SnoozedTabId: snoozedAt,
		SubEvent:     events.SubEventDelete,
	})

	err = h.notificationQueue.AddMessage(event)

	if err != nil {
		http_api.ErrorRes(w, errMsg.snoozedTabsCreate, http.StatusBadGateway)
		return
	}

	http_api.SuccessResMsg(w, "snoozed tab deleted successfully")
}

//* helpers

// middleware to get userId from jwt token present in req cookies
func newUserIdMiddleware() http_api.Handler {
	return func(w http.ResponseWriter, r *http.Request) {

		// get userId from jwt token
		userId := r.Header.Get("UserId")

		if userId == "" {
			http.Redirect(w, r, "/logout", http.StatusTemporaryRedirect)
			w.Header().Add("Error", "userId not found")
			return
		}

		r.SetPathValue("userId", userId)
	}
}
