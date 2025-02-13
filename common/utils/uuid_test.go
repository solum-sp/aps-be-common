package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUUID(t *testing.T) {
	var u UUID

	t.Run("New", func(t *testing.T) {
		uuid1 := u.New()
		uuid2 := u.New()
		assert.NotEqual(t, uuid1, uuid2, "New UUIDs should be unique")
		assert.NotEmpty(t, uuid1.String(), "UUID string should not be empty")
	})

	t.Run("String", func(t *testing.T) {
		id := u.New()
		str := id.String()
		assert.Len(t, str, 36, "UUID string should be 36 characters")
		assert.Contains(t, str, "-", "UUID string should contain hyphens")
	})

	t.Run("Parse", func(t *testing.T) {
		tests := []struct {
			name    string
			input   string
			wantErr bool
		}{
			{
				name:    "Valid UUID",
				input:   "123e4567-e89b-12d3-a456-426614174000",
				wantErr: false,
			},
			{
				name:    "Invalid UUID",
				input:   "invalid-uuid",
				wantErr: true,
			},
			{
				name:    "Empty string",
				input:   "",
				wantErr: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				parsed, err := u.Parse(tt.input)
				if tt.wantErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
					assert.Equal(t, tt.input, parsed.String())
				}
			})
		}
	})

	t.Run("MustParse", func(t *testing.T) {
		validUUID := "123e4567-e89b-12d3-a456-426614174000"
		assert.NotPanics(t, func() {
			parsed := u.MustParse(validUUID)
			assert.Equal(t, validUUID, parsed.String())
		})

		assert.Panics(t, func() {
			u.MustParse("invalid-uuid")
		})
	})
}
