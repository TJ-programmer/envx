package core

import "regexp"

var sensitiveNamePattern = regexp.MustCompile(`(?i)(SECRET|TOKEN|PASSWORD|PASSWD|PRIVATE_KEY|ACCESS_KEY|_KEY)`)

func IsSensitiveName(name string) bool {
	return sensitiveNamePattern.MatchString(name)
}
