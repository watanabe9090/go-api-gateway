package internal

import "testing"

func Test_GIVEN_ValidUserAndRole_THEN_Ok(t *testing.T) {
	tokenStr, err := CreateJwtToken("foobar", "USER", "a_secret_key")
	if err != nil {
		t.Error(err.Error())
	}
	token, err := ValidJwtToken(tokenStr, "a_secret_key")
	if err != nil {
		t.Error(err.Error())
	}
	sub, err := token.GetSubject()
	if err != nil {
		t.Error(err.Error())
	}
	if sub != "foobar" {
		t.Error("sub does not match the username")
	}
	aud, err := token.GetAudience()
	if err != nil {
		t.Error(err.Error())
	}
	if aud[0] != "USER" {
		t.Error("sub does not match the username")
	}
}

func Test_GIVEN_ValidUserAndRole_WHEN_SecretKeyIsEmpty_THEN_Ok(t *testing.T) {
	tokenStr, err := CreateJwtToken("foobar", "USER", "")
	if err != nil {
		t.Error(err.Error())
	}
	_, err = ValidJwtToken(tokenStr, "a_secret_key")
	if err == nil {
		t.Error("shold give signature error")
	}
}
