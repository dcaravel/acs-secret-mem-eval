package analyze

import (
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

func TestCalcSizeInBytes(t *testing.T) {
	// Get some baseline data
	var m map[string]DockerConfigEntry
	sizeOfNilMap := int(unsafe.Sizeof(m))
	assert.Equal(t, 8, sizeOfNilMap, "nil map")

	sizeOfNilString := int(unsafe.Sizeof(""))
	assert.Equal(t, 16, sizeOfNilString, "nil string")

	// size does not increase based on contents of string
	// true size will be the Sizeof + len(string)
	sizeOf123String := int(unsafe.Sizeof("123"))
	assert.Equal(t, 16, sizeOf123String, "123 string")

	var e DockerConfigEntry
	sizeOfNilEntry := int(unsafe.Sizeof(e))
	assert.Equal(t, sizeOfNilString*3, sizeOfNilEntry, "nil docker entry")

	m = map[string]DockerConfigEntry{}

	m["1"] = DockerConfigEntry{}
	assert.Equal(t, sizeOfNilEntry, calcSizeInBytes(m))

	m["1"] = DockerConfigEntry{Username: "123"}
	assert.Equal(t, sizeOfNilEntry+3, calcSizeInBytes(m))

	m["1"] = DockerConfigEntry{Username: "1", Password: "1", Email: "1"}
	assert.Equal(t, sizeOfNilEntry+3, calcSizeInBytes(m))

	m["2"] = DockerConfigEntry{Username: "1", Password: "1", Email: "1"}
	assert.Equal(t, (sizeOfNilEntry+3)*2, calcSizeInBytes(m))
}
