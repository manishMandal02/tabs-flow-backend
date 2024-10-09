package notes

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/manishMandal02/tabsflow-backend/pkg/http_api"
	"github.com/manishMandal02/tabsflow-backend/pkg/logger"
)

type noteHandler struct {
	nr noteRepository
}

func newNoteHandler(nr noteRepository) *noteHandler {
	return &noteHandler{
		nr: nr,
	}
}

func (h noteHandler) createNote(w http.ResponseWriter, r *http.Request) {
	userId := r.URL.Query().Get("userId")

	note := &note{}

	err := json.NewDecoder(r.Body).Decode(note)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = note.validate()

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.nr.createNote(userId, note)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http_api.SuccessResMsg(w, "Note created successfully")
}

func (h noteHandler) getNotes(w http.ResponseWriter, r *http.Request) {
	userId := r.URL.Query().Get("userId")

	lastKey := r.URL.Query().Get("lastNoteId")

	if lastKey == "" {
		lastKey = "0"
	}

	lastNoteId, err := strconv.ParseInt(lastKey, 10, 64)

	if err != nil {
		logger.Error("Couldn't parse lastNoteId", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	notes, err := h.nr.getNotes(userId, lastNoteId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http_api.SuccessResData(w, notes)

}

func (h noteHandler) updateNote(w http.ResponseWriter, r *http.Request) {
	userId := r.URL.Query().Get("userId")

	note := &note{}

	err := json.NewDecoder(r.Body).Decode(note)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = note.validate()

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.nr.updateNote(userId, note)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http_api.SuccessResMsg(w, "Note updated successfully")

}

func (h noteHandler) deleteNote(w http.ResponseWriter, r *http.Request) {
	userId := r.URL.Query().Get("userId")
	noteIdStr := r.URL.Query().Get("noteId")

	noteId, err := strconv.ParseInt(noteIdStr, 10, 64)

	if err != nil {
		logger.Error("Couldn't parse noteId", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.nr.deleteNote(userId, noteId)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http_api.SuccessResMsg(w, "Note deleted successfully")

}
