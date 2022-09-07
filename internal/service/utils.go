package service

import (
	"strings"

	"github.com/Dsmit05/img-loader/internal/models"
)

func CreateFileName(event models.UserEvent) string {
	var b strings.Builder
	b.WriteString(event.FirstName)
	b.WriteString("_")
	b.WriteString(event.LastName)

	return b.String()
}
