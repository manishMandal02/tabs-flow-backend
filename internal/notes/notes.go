package notes

import "github.com/go-playground/validator/v10"

type Note struct {
	Id          string `json:"id" validate:"required"`
	Title       string `json:"title" validate:"required"`
	Text        string `json:"text,omitempty" validate:"required"`
	SpaceId     string `json:"spaceId,omitempty"`
	Domain      string `json:"domain,omitempty"`
	RemainderAt int64  `json:"remainderAt,omitempty"`
	UpdatedAt   int64  `json:"updatedAt,omitempty"`
}

func (n *Note) validate() error {
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
	noteId           string
	notesGet         string
	notesGetEmpty    string
	noteDelete       string
	notesSearch      string
	notesSearchEmpty string
}{
	noteCreate:       "error creating note",
	noteUpdate:       "error updating note",
	noteId:           "note id is required",
	noteGet:          "error getting note",
	notesGetEmpty:    "notes not found",
	notesGet:         "error getting notes",
	noteDelete:       "error deleting note",
	notesSearch:      "error searching notes",
	notesSearchEmpty: "no notes found",
}
