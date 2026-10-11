package database

import (
	"errors"
	"fmt"
	"testing"

	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

func TestIsUniqueViolation(t *testing.T) {
	assert.True(t, IsUniqueViolation(&pq.Error{Code: "23505"}))
	assert.True(t, IsUniqueViolation(fmt.Errorf("wrapped: %w", &pq.Error{Code: "23505"})))
	assert.False(t, IsUniqueViolation(&pq.Error{Code: "23503"}))
	assert.False(t, IsUniqueViolation(errors.New("plain error")))
	assert.False(t, IsUniqueViolation(nil))
}

func TestIsForeignKeyViolation(t *testing.T) {
	assert.True(t, IsForeignKeyViolation(&pq.Error{Code: "23503"}))
	assert.False(t, IsForeignKeyViolation(&pq.Error{Code: "23505"}))
	assert.False(t, IsForeignKeyViolation(nil))
}
