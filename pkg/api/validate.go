package api

import "regexp"

var (
	registryRegexp = regexp.MustCompile(`^[0-9A-Za-z.-]+$`)
	repoRegexp     = regexp.MustCompile(`^[a-z0-9]+(?:[._-][a-z0-9]+)*$`)
	tagRegexp      = regexp.MustCompile(`^[\w][\w.-]{0,127}$`)
	cveRegexp      = regexp.MustCompile(`^[A-Z]+\-[\dA-Z]+\-\d+$`)
	githubRegexp   = regexp.MustCompile(`^https://(github.com|ghe.ckpd.co)/[0-9a-zA-Z-_]+/[0-9a-zA-Z-_]+$`)
)

func isValidRegistry(x string) bool {
	return registryRegexp.MatchString(x)
}

func isValidRepository(x string) bool {
	return repoRegexp.MatchString(x)
}

func isValidTag(x string) bool {
	return tagRegexp.MatchString(x)
}

func isValidGithubURL(x string) bool {
	return githubRegexp.MatchString(x)
}
