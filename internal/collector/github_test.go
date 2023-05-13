package collector

import "testing"

func TestMatchBot(t *testing.T) {
	var tests map[string]bool = map[string]bool{
		"test":           false,
		"testbot":        false,
		"test-bot":       true,
		"test[bot]":      true,
		"ChiaAutomation": true,
	}
	for testName, expect := range tests {
		t.Log(testName)
		result := matchesBot(testName)
		if result != expect {
			t.Errorf("Result fail for name %s", testName)
		}
	}
}
