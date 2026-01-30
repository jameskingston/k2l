package match

import (
	"strings"
)

// Match checks if the query matches the target using multiple strategies:
// 1. Exact match
// 2. Substring match
// 3. Dash-part match (e.g., "w-f" matches "web-frontend")
func Match(query, target string) bool {
	query = strings.ToLower(query)
	target = strings.ToLower(target)

	// Exact match
	if query == target {
		return true
	}

	// Substring match
	if strings.Contains(target, query) {
		return true
	}

	// Dash-part match: if query contains dashes, check if each part matches
	// corresponding parts in the target
	if strings.Contains(query, "-") && strings.Contains(target, "-") {
		return matchDashParts(query, target)
	}

	return false
}

// matchDashParts checks if each dash-separated part of the query
// matches (as substring) the corresponding part in the target.
// e.g., "w-f" matches "web-frontend", "api-s" matches "api-server"
func matchDashParts(query, target string) bool {
	queryParts := strings.Split(query, "-")
	targetParts := strings.Split(target, "-")

	if len(queryParts) > len(targetParts) {
		return false
	}

	for i, qp := range queryParts {
		if i >= len(targetParts) {
			return false
		}
		if !strings.Contains(targetParts[i], qp) {
			return false
		}
	}

	return true
}

// MatchResult contains the result of a match operation
type MatchResult struct {
	Match     string   // The best match (if unambiguous)
	Matches   []string // All matches at the same priority level
	Found     bool     // Whether any match was found
	Ambiguous bool     // Whether multiple matches exist at the same priority
}

// FindBestMatch returns the best matching item from a list.
// Priority: exact match > prefix match > substring match > dash-part match
// If multiple items match at the same priority level, returns all of them as ambiguous.
func FindBestMatch(query string, items []string) (string, bool) {
	result := FindMatches(query, items)
	if !result.Found {
		return "", false
	}
	return result.Match, true
}

// FindMatches returns detailed match results including ambiguity detection.
// Priority: exact match > prefix match > substring match > dash-part match
func FindMatches(query string, items []string) MatchResult {
	query = strings.ToLower(query)

	// First pass: exact match
	var exactMatches []string
	for _, item := range items {
		if strings.ToLower(item) == query {
			exactMatches = append(exactMatches, item)
		}
	}
	if len(exactMatches) == 1 {
		return MatchResult{Match: exactMatches[0], Matches: exactMatches, Found: true, Ambiguous: false}
	}
	if len(exactMatches) > 1 {
		return MatchResult{Match: exactMatches[0], Matches: exactMatches, Found: true, Ambiguous: true}
	}

	// Second pass: prefix match
	var prefixMatches []string
	for _, item := range items {
		if strings.HasPrefix(strings.ToLower(item), query) {
			prefixMatches = append(prefixMatches, item)
		}
	}
	if len(prefixMatches) == 1 {
		return MatchResult{Match: prefixMatches[0], Matches: prefixMatches, Found: true, Ambiguous: false}
	}
	if len(prefixMatches) > 1 {
		return MatchResult{Match: prefixMatches[0], Matches: prefixMatches, Found: true, Ambiguous: true}
	}

	// Third pass: substring match
	var substringMatches []string
	for _, item := range items {
		if strings.Contains(strings.ToLower(item), query) {
			substringMatches = append(substringMatches, item)
		}
	}
	if len(substringMatches) == 1 {
		return MatchResult{Match: substringMatches[0], Matches: substringMatches, Found: true, Ambiguous: false}
	}
	if len(substringMatches) > 1 {
		return MatchResult{Match: substringMatches[0], Matches: substringMatches, Found: true, Ambiguous: true}
	}

	// Fourth pass: dash-part match
	var dashMatches []string
	for _, item := range items {
		if strings.Contains(query, "-") && strings.Contains(item, "-") {
			if matchDashParts(query, strings.ToLower(item)) {
				dashMatches = append(dashMatches, item)
			}
		}
	}
	if len(dashMatches) == 1 {
		return MatchResult{Match: dashMatches[0], Matches: dashMatches, Found: true, Ambiguous: false}
	}
	if len(dashMatches) > 1 {
		return MatchResult{Match: dashMatches[0], Matches: dashMatches, Found: true, Ambiguous: true}
	}

	return MatchResult{Found: false}
}

// FindAllMatches returns all items that match the query
func FindAllMatches(query string, items []string) []string {
	var matches []string
	for _, item := range items {
		if Match(query, item) {
			matches = append(matches, item)
		}
	}
	return matches
}
