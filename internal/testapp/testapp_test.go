package testapp

import "testing"

func TestHelloWorld(t *testing.T) {
	want := "Hello, Robot Warehouse!"
	got := HelloWorld()
	if got != want {
		t.Errorf("Expected %q but got %q", want, got)
	}
}
