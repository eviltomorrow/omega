package lock

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateFile(t *testing.T) {
	_assert := assert.New(t)
	lock, err := CreateFileLock("test.pid")
	_assert.Nil(err)
	_assert.NotNil(lock)

	lock, err = CreateFileLock("test.pid")
	_assert.NotNil(err)
	_assert.Nil(lock)

	err = DestroyFileLock(lock)
	_assert.Nil(err)
}
