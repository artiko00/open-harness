package svc

import "testing"

func TestPing(t *testing.T) {
	if Ping() != "pong" {
		t.Fail()
	}
}
