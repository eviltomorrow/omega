package lock

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewFileLock(t *testing.T) {
	_assert := assert.New(t)
	lock, err := newFileLock("testfile.lock", false)
	_assert.Nil(err)
	_assert.NotNil(lock)

	lock, err = newFileLock("testfile.lock", false)
	_assert.NotNil(err)
	_assert.Nil(lock)
}
