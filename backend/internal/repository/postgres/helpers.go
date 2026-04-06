package postgres

import (
	"strings"

	"workflow-engine/backend/internal/repository"
)

func wrapDBError(err error) error {
	if err == nil {
		return nil
	}
	msg := err.Error()
	if strings.Contains(msg, "violates foreign key constraint") {
		return repository.ErrForeignKey
	}
	return err
}
