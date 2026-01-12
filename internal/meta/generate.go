package meta

import (
	"github.com/samakintunde/bujo-cli/internal/id"
	"github.com/samakintunde/bujo-cli/internal/models"
)

func Generate() models.Metadata {
	return models.Metadata{
		ID:   id.New(),
		Mig:  0,
		PID:  "",
		Rsch: 0,
	}
}
