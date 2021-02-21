package sim

import "testing"

func assertEqual(t *testing.T, a interface{}, b interface{}, msg string) {
	if a != b {
		t.Fatalf("%s: %s != %s", msg, a, b)
	}
}
