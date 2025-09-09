package project

import (
	"time"

	"github.com/gocasters/rankr/projectapp/constant"
)

type ProjectEntity struct {
	ID                 string                 `db:"id" json:"id"`
	Name               string                 `db:"name" json:"name"`
	Slug               string                 `db:"slug" json:"slug"`
	Description        *string                `db:"description" json:"description,omitempty"`
	DesignReferenceURL *string                `db:"design_reference_url" json:"designReferenceUrl,omitempty"`
	Status             constant.ProjectStatus `db:"status" json:"status"`
	CreatedAt          time.Time              `db:"created_at" json:"createdAt"`
	UpdatedAt          time.Time              `db:"updated_at" json:"updatedAt"`
	ArchivedAt         *time.Time             `db:"archived_at" json:"archivedAt,omitempty"`
}
