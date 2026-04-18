// Package migrations runs GORM AutoMigrate to keep the schema in sync.
// For production use, replace AutoMigrate with a proper migration tool (goose, atlas).
package migrations

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/dentdesk/backend/internal/domain"
)

// Run executes AutoMigrate for all domain models in dependency order.
func Run(db *gorm.DB) error {
	models := []any{
		&domain.Clinic{},
		&domain.User{},
		&domain.Doctor{},
		&domain.Patient{},
		&domain.Appointment{},
		&domain.MessageLog{},
	}

	for _, m := range models {
		if err := db.AutoMigrate(m); err != nil {
			return fmt.Errorf("auto-migrate %T: %w", m, err)
		}
	}

	return nil
}
