package utils

import (
	"uuid"
)

func GenerateUUID() string {
	idV4 := uuid.NewV4()
	return idV4.String()

}
