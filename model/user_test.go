package model

import "testing"

func TestSetPassowrdAndComparePassword(t *testing.T) {
	user := User{}

	plainPassword := "letstest"
	user.SetPassword(plainPassword)

	if !user.IsPasswordValid(plainPassword) {
		t.Error("expect valid password but got invalid")
	}
}
