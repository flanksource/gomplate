package funcs

import (
	"context"
	"net/url"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateUUIDFuncs(t *testing.T) {
	t.Parallel()

	for i := 0; i < 10; i++ {
		// Run this a bunch to catch race conditions
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			fmap := CreateUUIDFuncs(ctx)
			actual := fmap["uuid"].(func() interface{})

			assert.Equal(t, ctx, actual().(*UUIDFuncs).ctx)
		})
	}
}

const (
	uuidV1Pattern = "^[[:xdigit:]]{8}-[[:xdigit:]]{4}-1[[:xdigit:]]{3}-[89ab][[:xdigit:]]{3}-[[:xdigit:]]{12}$"
	uuidV4Pattern = "^[[:xdigit:]]{8}-[[:xdigit:]]{4}-4[[:xdigit:]]{3}-[89ab][[:xdigit:]]{3}-[[:xdigit:]]{12}$"
)

func TestV1(t *testing.T) {
	t.Parallel()

	u := UUIDNS()
	i, err := u.V1()
	assert.NoError(t, err)
	assert.Regexp(t, uuidV1Pattern, i)
}

func TestV4(t *testing.T) {
	t.Parallel()

	u := UUIDNS()
	i, err := u.V4()
	assert.NoError(t, err)
	assert.Regexp(t, uuidV4Pattern, i)
}

func TestNil(t *testing.T) {
	t.Parallel()

	u := UUIDNS()
	i, err := u.Nil()
	assert.NoError(t, err)
	assert.Equal(t, "00000000-0000-0000-0000-000000000000", i)
}

func TestIsValid(t *testing.T) {
	t.Parallel()

	u := UUIDNS()
	in := interface{}(false)
	i, err := u.IsValid(in)
	assert.NoError(t, err)
	assert.False(t, i)

	in = 12345
	i, err = u.IsValid(in)
	assert.NoError(t, err)
	assert.False(t, i)

	testdata := []interface{}{
		"123456781234123412341234567890ab",
		"12345678-1234-1234-1234-1234567890ab",
		"urn:uuid:12345678-1234-1234-1234-1234567890ab",
		"{12345678-1234-1234-1234-1234567890ab}",
	}

	for _, d := range testdata {
		i, err = u.IsValid(d)
		assert.NoError(t, err)
		assert.True(t, i)
	}
}

func TestParse(t *testing.T) {
	t.Parallel()

	u := UUIDNS()
	in := interface{}(false)
	_, err := u.Parse(in)
	assert.Error(t, err)

	in = 12345
	_, err = u.Parse(in)
	assert.Error(t, err)

	in = "12345678-1234-1234-1234-1234567890ab"
	testdata := []interface{}{
		"123456781234123412341234567890ab",
		"12345678-1234-1234-1234-1234567890ab",
		"urn:uuid:12345678-1234-1234-1234-1234567890ab",
		must(url.Parse("urn:uuid:12345678-1234-1234-1234-1234567890ab")),
		"{12345678-1234-1234-1234-1234567890ab}",
	}

	for _, d := range testdata {
		uid, err := u.Parse(d)
		assert.NoError(t, err)
		assert.Equal(t, in, uid)
	}
}

func TestIdempotentUUID(t *testing.T) {
	t.Parallel()

	u := UUIDNS()
	
	// Test that the same input produces the same UUID
	uuid1, err := u.IdempotentUUID("test")
	assert.NoError(t, err)
	assert.NotEmpty(t, uuid1)
	
	uuid2, err := u.IdempotentUUID("test")
	assert.NoError(t, err)
	assert.Equal(t, uuid1, uuid2, "Same input should produce same UUID")
	
	// Test that different inputs produce different UUIDs
	uuid3, err := u.IdempotentUUID("different")
	assert.NoError(t, err)
	assert.NotEqual(t, uuid1, uuid3, "Different inputs should produce different UUIDs")
	
	// Test with multiple arguments
	uuid4, err := u.IdempotentUUID("arg1", "arg2", "arg3")
	assert.NoError(t, err)
	assert.NotEmpty(t, uuid4)
	
	uuid5, err := u.IdempotentUUID("arg1", "arg2", "arg3")
	assert.NoError(t, err)
	assert.Equal(t, uuid4, uuid5, "Same multiple arguments should produce same UUID")
	
	// Test that order matters
	uuid6, err := u.IdempotentUUID("arg2", "arg1", "arg3")
	assert.NoError(t, err)
	assert.NotEqual(t, uuid4, uuid6, "Different order should produce different UUIDs")
	
	// Test with numeric arguments
	uuid7, err := u.IdempotentUUID(123, 456)
	assert.NoError(t, err)
	assert.NotEmpty(t, uuid7)
	
	uuid8, err := u.IdempotentUUID(123, 456)
	assert.NoError(t, err)
	assert.Equal(t, uuid7, uuid8, "Same numeric arguments should produce same UUID")
	
	// Test with no arguments
	uuid9, err := u.IdempotentUUID()
	assert.NoError(t, err)
	assert.NotEmpty(t, uuid9)
	
	uuid10, err := u.IdempotentUUID()
	assert.NoError(t, err)
	assert.Equal(t, uuid9, uuid10, "No arguments should produce same UUID each time")
	
	// Verify the UUID format is valid (version 5, SHA-based)
	parsed, err := u.Parse(uuid1)
	assert.NoError(t, err)
	assert.Equal(t, uuid1, parsed)
	
	// Check that it's a valid UUID v5 pattern
	assert.Regexp(t, "^[[:xdigit:]]{8}-[[:xdigit:]]{4}-5[[:xdigit:]]{3}-[89ab][[:xdigit:]]{3}-[[:xdigit:]]{12}$", uuid1)
}
