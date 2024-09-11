package internal

import "testing"

func Test_WHEN_ValidatingHashFromSameString_THEN_Ok(t *testing.T) {
	hash, err := HashPassword("some_secret_of_foobar")
	if err != nil {
		t.Error(err.Error())
	}
	if !ValidatePasswordHash(hash, "some_secret_of_foobar") {
		t.Error("hash is not matching same string")
	}
}
