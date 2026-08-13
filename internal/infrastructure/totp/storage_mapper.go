package totp

import (
	"encoding/base32"
	"net/url"
	"strconv"
	"time"
)

// FieldName is the custom KDBX entry field this application reads and
// writes for TOTP configuration: the full otpauth:// URI, stored as one
// string field. This matches KeePassXC's own modern convention — its
// TOTP settings dialog reads and writes a custom attribute literally
// named "otp" holding the URI — which Strongbox/KeePassium and
// KeePassDX/Keepass2Android have also added support for reading,
// alongside the older two-field "TOTP Seed"/"TOTP Settings" convention
// this application does not write (spec section 14.3: use the most
// widespread format when a single one can't reach every target client).
//
// Spec 14.3 also requires this choice to be backed by a fixed
// interoperability matrix tested against the real reference clients
// before TOTP write support ships. That verification hasn't happened in
// this environment — this constant names the convention the research is
// expected to confirm, not a substitute for doing it.
const FieldName = "otp"

// BuildURI serializes params (plus label and issuer) into the
// otpauth://totp/... form ParseURI accepts, so the two round-trip:
// ParseURI(BuildURI(label, issuer, params)) reproduces params. Every
// parameter is written explicitly — including ones equal to RFC 6238's
// defaults — rather than omitted, so a reader never has to guess this
// application's defaults match its own.
func BuildURI(label, issuer string, params Params) string {
	period := params.Period
	if period <= 0 {
		period = 30 * time.Second
	}

	q := url.Values{}
	q.Set("secret", base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(params.Secret))
	if issuer != "" {
		q.Set("issuer", issuer)
	}
	q.Set("algorithm", string(params.Algorithm))
	q.Set("digits", strconv.Itoa(params.Digits))
	q.Set("period", strconv.Itoa(int(period.Seconds())))

	u := url.URL{
		Scheme:   "otpauth",
		Host:     "totp",
		Path:     "/" + label,
		RawQuery: q.Encode(),
	}
	return u.String()
}
