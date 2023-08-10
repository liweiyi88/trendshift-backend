package model

import "testing"

func TestSetPassowrdAndComparePassword(t *testing.T) {
	user := User{}

	plainPassword := "letstest"
	user.setPassword(plainPassword)

	if !user.isPasswordValid(plainPassword) {
		t.Error("expect valid password but got invalid")
	}
}
