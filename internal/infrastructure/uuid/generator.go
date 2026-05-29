package uuid

import (
	"mizu/internal/domain/common"

	"github.com/google/uuid"
)

type UUIDGenerator struct{}

func NewUUIDGenerator() *UUIDGenerator {
	return &UUIDGenerator{}
}

func (*UUIDGenerator) Generate() (string, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return "", common.ErrFailedToGenerateID
	}

	return id.String(), nil
}
