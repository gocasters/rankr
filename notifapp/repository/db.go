package repository

import (
	"github.com/gocasters/rankr/pkg/database"
)

func New(db *database.Database) Repository {
	return Repository{db: db}
}

type Repository struct {
	db *database.Database
}
