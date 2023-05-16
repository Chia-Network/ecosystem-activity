package utils

import "strings"

// Bots is used to match usernames to likely bots, and skips their commit data
var Bots = []string{
	"-bot",
	"[bot]",
	"ChiaAutomation",
	"deepsourcebot",
}

// MatchesBot matches common bot names since we don't care for bot users in our data
func MatchesBot(uname string) bool {
	for _, botMatcher := range Bots {
		if strings.Contains(strings.ToLower(uname), strings.ToLower(botMatcher)) {
			return true
		}
	}
	return false
}
