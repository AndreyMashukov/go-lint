package a

import "testing"

type tFake struct{}

func (tFake) Errorf(string, ...interface{}) {}

type fakeAssert struct{}

func (fakeAssert) NotNil(t interface{}, v interface{}) bool     { return true }
func (fakeAssert) IsType(t interface{}, e, v interface{}) bool  { return true }
func (fakeAssert) NotEmpty(t interface{}, v interface{}) bool   { return true }
func (fakeAssert) NotZero(t interface{}, v interface{}) bool    { return true }
func (fakeAssert) Implements(t, i, v interface{}) bool          { return true }
func (fakeAssert) Equal(t, e, v interface{}) bool               { return true }

var assert = fakeAssert{}
var require = fakeAssert{}

func TestBad(t *testing.T) {
	var x interface{} = "hi"
	assert.NotNil(t, x)     // want `type-only or existence-only assertion`
	assert.IsType(t, "", x) // want `type-only or existence-only assertion`
	require.NotEmpty(t, x)  // want `type-only or existence-only assertion`
	assert.NotZero(t, x)    // want `type-only or existence-only assertion`
	assert.Implements(t, (*error)(nil), x) // want `type-only or existence-only assertion`
	assert.Equal(t, "hi", x)
}
