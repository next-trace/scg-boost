package dbquery

import (
	"regexp"
	"strings"
)

// IsReadOnly checks if the query is read-only (SELECT or WITH CTE).
// It strips comments, rejects multiple statements, and blocks write keywords.
func IsReadOnly(query string) bool {
	// Strip comments
	re := regexp.MustCompile(`--.*`)
	query = re.ReplaceAllString(query, "")
	re = regexp.MustCompile(`/\*.*?\*/`)
	query = re.ReplaceAllString(query, "")
	query = strings.TrimSpace(query)

	// Reject multiple statements
	if strings.Contains(query, ";") {
		parts := strings.Split(query, ";")
		if len(parts) > 1 && strings.TrimSpace(parts[1]) != "" {
			return false
		}
		query = parts[0]
	}

	// Block write keywords
	blocked := regexp.MustCompile(`(?i)\b(INSERT|UPDATE|DELETE|MERGE|ALTER|TRUNCATE|DROP|CREATE|REINDEX|VACUUM|ANALYZE|GRANT|REVOKE|COPY|CALL|DO)\b`)
	if blocked.MatchString(query) {
		return false
	}

	// Allow SELECT or WITH
	allowed := regexp.MustCompile(`(?i)^\s*(SELECT|WITH)\b`)
	return allowed.MatchString(query)
}
