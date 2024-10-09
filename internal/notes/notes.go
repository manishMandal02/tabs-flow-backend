package notes

import "github.com/go-playground/validator/v10"

type note struct {
	Id          string `json:"id"`
	Title       string `json:"title"`
	Text        string `json:"text"`
	SpaceId     string `json:"spaceId"`
	Domain      string `json:"domain"`
	RemainderAt int64  `json:"remainderAt"`
	CreatedAt   int64  `json:"createdAt"`
	UpdatedAt   int64  `json:"updatedAt"`
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
	noteCreate string
	noteUpdate string
	noteGet    string
	notesGet   string
	noteDelete string
}{
	noteCreate: "error creating note",
	noteUpdate: "error updating note",
	noteGet:    "error getting note",
	notesGet:   "error getting notes",
	noteDelete: "error deleting note",
}
