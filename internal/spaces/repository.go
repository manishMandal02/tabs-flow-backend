package spaces

import "github.com/manishMandal02/tabsflow-backend/pkg/database"

type spaceRepository interface {
	createSpace(s *space) error
	getSpaceById(id string) (*space, error)
	updateSpace(id string, s *space) error
	deleteSpace(id string) error
}

type spaceRepo struct {
	db *database.DDB
}

func newSpaceRepository(db *database.DDB) spaceRepository {
	return &spaceRepo{
		db: db,
	}
}

func (r *spaceRepo) createSpace(s *space) error {
	return nil
}

func (r *spaceRepo) getSpaceById(id string) (*space, error) {
	return nil, nil
}

func (r *spaceRepo) updateSpace(id string, s *space) error {
	return nil
}

func (r *spaceRepo) deleteSpace(id string) error {
	return nil
}
