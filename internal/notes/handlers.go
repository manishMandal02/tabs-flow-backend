package notes

import (
	"encoding/json"
	"net/http"

	"github.com/manishMandal02/tabsflow-backend/pkg/http_api"
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

	notes, err := h.nr.getNotes(userId)
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
	noteId := r.URL.Query().Get("noteId")

	err := h.nr.deleteNote(userId, noteId)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http_api.SuccessResMsg(w, "Note deleted successfully")

}
