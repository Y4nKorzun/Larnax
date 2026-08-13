// Package googlecsv implements the Google Password Manager CSV importer
// (spec section 13).
package googlecsv

import (
	"net/url"
	"strings"
)

// normalizeOrigin reduces rawURL to a scheme+host+port identity for
// duplicate comparison (spec section 13.6's "normalized origin"). Path,
// query, and fragment are deliberately discarded: two entries for the same
// site's login page and its account-settings page should still be flagged
// as possible duplicates.
//
// Normalization:
//   - scheme and host are lowercased (both are case-insensitive per RFC 3986);
//   - a missing scheme defaults to "https", the overwhelmingly common case
//     for a bare host pasted or exported without one;
//   - a leading "www." on the host is stripped, since it almost always
//     identifies the same login realm as the bare domain;
//   - the scheme's default port (80 for http, 443 for https) is dropped,
//     since an explicit ":443" and no port at all mean the same thing.
//
// An empty or unparseable rawURL normalizes to "".
func normalizeOrigin(rawURL string) string {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return ""
	}
	if !strings.Contains(rawURL, "://") {
		rawURL = "https://" + rawURL
	}

	u, err := url.Parse(rawURL)
	if err != nil || u.Host == "" {
		return ""
	}

	scheme := strings.ToLower(u.Scheme)
	host := strings.ToLower(u.Hostname())
	host = strings.TrimPrefix(host, "www.")
	if host == "" {
		return ""
	}

	if port := u.Port(); port != "" && port != defaultPortForScheme(scheme) {
		return scheme + "://" + host + ":" + port
	}
	return scheme + "://" + host
}

func defaultPortForScheme(scheme string) string {
	switch scheme {
	case "http":
		return "80"
	case "https":
		return "443"
	default:
		return ""
	}
}

// normalizeUsername reduces username to a form suitable for duplicate
// comparison (spec section 13.6): trimmed and lowercased. Most login
// usernames are email addresses or handles that are conventionally
// case-insensitive in practice, even where case sensitivity is technically
// implementation-defined.
func normalizeUsername(username string) string {
	return strings.ToLower(strings.TrimSpace(username))
}
