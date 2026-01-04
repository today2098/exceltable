package exceltable

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegisterDefaultPredicate(t *testing.T) {
	predFunc := func() bool { return true }
	err := RegisterDefaultPredicate("test", predFunc) // nullary predicate
	assert.NoError(t, err)

	storedPred, ok := defaultPredicates.load("test")
	assert.True(t, ok)
	assert.Equal(t, reflect.ValueOf(predFunc).Type(), storedPred.typ)

	err = RegisterDefaultPredicate("test2", func(s string) bool { return s == "something" }) // unary predicate
	assert.NoError(t, err)

	_, ok = defaultPredicates.load("test2")
	assert.True(t, ok)

	err = RegisterDefaultPredicate("test2", predFunc) // re-register by existing key
	assert.NoError(t, err)

	storedPred, ok = defaultPredicates.load("test2")
	assert.True(t, ok)
	assert.Equal(t, reflect.ValueOf(predFunc).Type(), storedPred.typ)

	err = RegisterDefaultPredicate("test3", func(a, b int) bool { return a == b }) // invalid predicate with two arguments
	assert.Error(t, err)

	_, ok = defaultPredicates.load("test3")
	assert.False(t, ok)

	err = RegisterDefaultPredicate("test4", func() int { return 1 }) // invalid predicate with non-bool return
	assert.Error(t, err)

	_, ok = defaultPredicates.load("test4")
	assert.False(t, ok)

	err = RegisterDefaultPredicate("test5", 123) // non-function predicate
	assert.Error(t, err)

	_, ok = defaultPredicates.load("test5")
	assert.False(t, ok)
}
