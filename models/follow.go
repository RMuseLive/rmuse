package models

import (
	"encoding/json"
	"time"

	"github.com/markbates/pop"
	"github.com/markbates/validate"
	"github.com/satori/go.uuid"
)

type Follow struct {
	ID        uuid.UUID `json:"id" db:"id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	Follower  uuid.UUID `json:"-"          db:"follower"`
	Followed  uuid.UUID `json:"-"          db:"followed"`
}

func (f *Follow) Create(tx *pop.Connection) (*validate.Errors, error) {
	return tx.ValidateAndCreate(f)
}

func (f *Follow) Delete(tx *pop.Connection) error {
	query := tx.RawQuery("DELETE FROM follows WHERE follower = ? AND followed = ?", f.Follower, f.Followed)
	return query.Exec()
}

// String is not required by pop and may be deleted
func (f Follow) String() string {
	jf, _ := json.Marshal(f)
	return string(jf)
}

// Follows is not required by pop and may be deleted
type Follows []Follow

// String is not required by pop and may be deleted
func (f Follows) String() string {
	jf, _ := json.Marshal(f)
	return string(jf)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (f *Follow) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
// This method is not required and may be deleted.
func (f *Follow) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
// This method is not required and may be deleted.
func (f *Follow) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}
