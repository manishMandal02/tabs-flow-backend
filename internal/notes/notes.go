package notes

import "github.com/go-playground/validator/v10"

type note struct {
	Id          string `json:"id"`
	Title       string `json:"title"`
	Text        string `json:"text"`
	SpaceId     string `json:"spaceId,omitempty"`
	Domain      string `json:"domain,omitempty"`
	RemainderAt int64  `json:"remainderAt,omitempty"`
	UpdatedAt   int64  `json:"updatedAt,omitempty"`
}

func (n *note) validate() error {
	validate := validator.New(validator.WithRequiredStructEnabled())
	err := validate.Struct(n)
	if err != nil {
		return err
	}

	return nil
}

var errMsg = struct {
	noteCreate       string
	noteUpdate       string
	noteGet          string
	notesGet         string
	notesGetEmpty    string
	noteDelete       string
	notesSearch      string
	notesSearchEmpty string
}{
	noteCreate:       "error creating note",
	noteUpdate:       "error updating note",
	noteGet:          "error getting note",
	notesGetEmpty:    "no notes found",
	notesGet:         "error getting notes",
	noteDelete:       "error deleting note",
	notesSearch:      "error searching notes",
	notesSearchEmpty: "no notes found",
}
