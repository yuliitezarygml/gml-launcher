package models

import (
	"encoding/json"
	"testing"
)

func TestUserSerialization(t *testing.T) {
	u := User{
		UUID:        "test-uuid",
		Login:       "testuser",
		Password:    "pass",
		IsSlim:      false,
		Blocked:     false,
		BlockReason: "",
	}
	data, err := json.Marshal(u)
	if err != nil {
		t.Fatal(err)
	}
	var u2 User
	if err := json.Unmarshal(data, &u2); err != nil {
		t.Fatal(err)
	}
	if u2.Login != "testuser" {
		t.Errorf("expected testuser, got %s", u2.Login)
	}
}
