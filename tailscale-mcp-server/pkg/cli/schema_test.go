package cli

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestStruct struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func TestParseSchema(t *testing.T) {
	t.Parallel()

	raw := `{"id": 1, "name": "test"}`
	dst, err := ParseSchema[TestStruct](raw)

	require.NoError(t, err)
	assert.Equal(t, 1, dst.ID)
	assert.Equal(t, "test", dst.Name)
}

func TestParseSchemaWithValidator(t *testing.T) {
	t.Parallel()

	raw := `{"id": 1, "name": "test"}`
	dst, err := ParseSchemaWithValidator(raw, func(dst TestStruct) error {
		if dst.ID < 0 {
			return errors.New("id must be positive")
		}
		return nil
	})

	require.NoError(t, err)
	assert.Equal(t, 1, dst.ID)
	assert.Equal(t, "test", dst.Name)

	raw = `{"id": -1, "name": "test"}`
	dst, err = ParseSchemaWithValidator(raw, func(dst TestStruct) error {
		if dst.ID < 0 {
			return errors.New("id must be positive")
		}
		return nil
	})
	require.Error(t, err)
}
