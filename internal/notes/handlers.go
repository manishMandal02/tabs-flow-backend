package notes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/kljensen/snowball"
	"github.com/manishMandal02/tabsflow-backend/pkg/http_api"
	"github.com/manishMandal02/tabsflow-backend/pkg/logger"
)

type noteHandler struct {
	r noteRepository
}

func newNoteHandler(nr noteRepository) *noteHandler {
	return &noteHandler{
		r: nr,
	}
}

func (h noteHandler) create(w http.ResponseWriter, r *http.Request) {
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

	err = h.r.createNote(userId, note)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// index search terms in search table
	terms :=
		extractSearchTerms(note.Title, note.Text, note.Domain)

	err = h.r.indexSearchTerms(userId, note.Id, terms)

	if err != nil {
		logger.Error(fmt.Sprintf("error indexing search terms for note: %v", note), err)
	}

	http_api.SuccessResMsg(w, "Note created successfully")
}

func (h noteHandler) get(w http.ResponseWriter, r *http.Request) {
	userId := r.URL.Query().Get("userId")

	noteId := r.URL.Query().Get("noteId")

	if noteId == "" {
		http.Error(w, errMsg.noteGet, http.StatusBadRequest)
		return
	}

	notes, err := h.r.getNote(userId, noteId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http_api.SuccessResData(w, notes)
}

func (h noteHandler) getAllByUser(w http.ResponseWriter, r *http.Request) {
	userId := r.URL.Query().Get("userId")
	lastNoteIdStr := r.URL.Query().Get("lastNoteId")

	lastNoteId, err := strconv.ParseInt(lastNoteIdStr, 10, 64)

	if err != nil {
		logger.Error("Couldn't parse noteId", err)
		http.Error(w, errMsg.noteGet, http.StatusBadRequest)
		return
	}

	note, err := h.r.getNotesByUser(userId, lastNoteId)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http_api.SuccessResData(w, note)

}

func (h noteHandler) search(w http.ResponseWriter, r *http.Request) {
	userId := r.URL.Query().Get("userId")
	query := r.URL.Query().Get("query")
	maxSearchLimit := r.URL.Query().Get("limit")

	searchTerms := strings.Fields(query)

	limit := 8

	if maxSearchLimit != "" {
		n, err := strconv.ParseInt(maxSearchLimit, 10, 32)
		if err != nil {
			logger.Error("Couldn't parse search limit query", err)
			http.Error(w, errMsg.notesSearch, http.StatusBadRequest)
			return
		}
		limit = int(n)
	}

	notesIds, err := getNoteIdsBySearchTerms(userId, searchTerms, limit, h.r)

	if err != nil {
		http.Error(w, errMsg.notesSearch, http.StatusInternalServerError)
		return
	}

	// get notes that matched the search query
	notes, err := h.r.getNotesByIds(userId, &notesIds)

	if err != nil {
		http.Error(w, errMsg.notesSearch, http.StatusInternalServerError)
		return
	}

	if len(*notes) == 0 {
		http.Error(w, errMsg.notesSearchEmpty, http.StatusBadGateway)
		return
	}

	http_api.SuccessResData(w, notes)
}

func (h noteHandler) update(w http.ResponseWriter, r *http.Request) {
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

	err = h.r.updateNote(userId, note)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// TODO: if title, note or domain is updated, re-index search terms

	http_api.SuccessResMsg(w, "Note updated successfully")

}

func (h noteHandler) delete(w http.ResponseWriter, r *http.Request) {
	userId := r.URL.Query().Get("userId")
	noteIdStr := r.URL.Query().Get("noteId")

	noteId, err := strconv.ParseInt(noteIdStr, 10, 64)

	if err != nil {
		logger.Error("Couldn't parse noteId", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.r.deleteNote(userId, noteId)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http_api.SuccessResMsg(w, "Note deleted successfully")

}

// helpers
func createSearchTermPK(userId string, term string) string {
	return fmt.Sprintf("%s#%s", userId, term)
}

func extractSearchTerms(title, note, domainName string) []string {
	allText := strings.ToLower(title + " " + note)

	words := strings.Fields(allText)

	stemmedTerms := make(map[string]bool)

	for _, word := range words {
		if len(word) < 3 || isCommonWord(word) {
			continue
		}
		stemmed, _ := snowball.Stem(word, "english", true)

		stemmedTerms[stemmed] = true

		return nil
	}

	searchTerms := []string{}

	for term := range stemmedTerms {
		searchTerms = append(searchTerms, term)
	}

	// add domain name as search terms
	if domainName != "" {

		// full domain as search term
		searchTerms = append(searchTerms, domainName)

		domainTerms := strings.Split(domainName, ".")

		// domain without extension as search term

		if len(domainTerms) < 3 {
			// no subdomain
			searchTerms = append(searchTerms, domainTerms[0])
		} else {
			withOutExt := domainTerms[:len(domainTerms)-1]
			domainName = strings.Join(withOutExt, ".")

			searchTerms = append(searchTerms, domainName)
		}
	}

	return searchTerms
}

func isCommonWord(word string) bool {
	commonWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true, "but": true,
		"in": true, "on": true, "at": true, "to": true, "for": true, "of": true,
		"with": true, "by": true, "from": true, "up": true, "about": true, "into": true,
		"over": true, "after": true, "is": true, "are": true, "was": true, "were": true,
	}

	return commonWords[word]
}

func getNoteIdsBySearchTerms(userId string, searchTerms []string, limit int, r noteRepository) ([]string, error) {

	noteIdSets := []map[string]bool{}

	for _, term := range searchTerms {
		stemmed, _ := snowball.Stem(term, "english", true)
		noteIds, err := r.findSearchIndex(userId, stemmed, limit)

		if err != nil {
			if err.Error() == errMsg.notesSearchEmpty {
				continue
			} else {
				return nil, err
			}
		}

		noteIdSet := make(map[string]bool)
		for _, id := range noteIds {
			noteIdSet[id] = true
		}

		noteIdSets = append(noteIdSets, noteIdSet)
	}

	// Find intersection of note IDs
	intersection := noteIdSets[0]
	for _, set := range noteIdSets[1:] {
		for id := range intersection {
			if !set[id] {
				delete(intersection, id)
			}
		}
	}

	notesIdsMatched := make([]string, 0, len(intersection))

	for id := range intersection {
		notesIdsMatched = append(notesIdsMatched, id)
	}

	notesIdsMatched = notesIdsMatched[:limit]

	return notesIdsMatched, nil
}
