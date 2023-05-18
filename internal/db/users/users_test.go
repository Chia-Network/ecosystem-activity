package users

import "testing"

func TestGetBotLikes(t *testing.T) {
	testBots := []string{
		"test",
		"test2",
		"test3",
	}
	expect := `username LIKE '%test%' OR username LIKE '%test2%' OR username LIKE '%test3%'`

	if result := getBotLikes(testBots); result != expect {
		t.Errorf("Result fail. Received %s, Expected %s", result, expect)
	}
}
