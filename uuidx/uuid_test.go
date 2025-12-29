package uuidx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsUUID(t *testing.T) {
	assert.True(t, IsUUID("f40a3f19-957c-4538-acf9-0b0b9c92db9c"))
	assert.False(t, IsUUID("f40a3f19-957c-4538-acf9-0b0b9c92db9"))
	assert.False(t, IsUUID("f40a3f19-957c-4538-acf9-0b0b9c92db9c1"))
	assert.False(t, IsUUID("f"))
}
