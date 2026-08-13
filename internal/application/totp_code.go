package application

import (
	"errors"
	"time"

	"github.com/Y4nKorzun/Larnax/internal/domain"
	"github.com/Y4nKorzun/Larnax/internal/infrastructure/totp"
)

// ErrNoTOTPField is returned by GenerateTOTPCode and TOTPTimeRemaining
// when entry has no TOTP custom field (totp.EntryURI/SetEntryURI) to
// read from.
var ErrNoTOTPField = errors.New("application: entry has no TOTP field")

// entryTOTPParams reads and parses entry's TOTP URI, the shared first
// step GenerateTOTPCode and TOTPTimeRemaining both need.
func entryTOTPParams(entry domain.Entry) (totp.Params, error) {
	uri, present, err := totp.EntryURI(entry)
	if err != nil {
		return totp.Params{}, err
	}
	if !present {
		return totp.Params{}, ErrNoTOTPField
	}
	parsed, err := totp.ParseURI(uri)
	if err != nil {
		return totp.Params{}, err
	}
	return parsed.Params, nil
}

// GenerateTOTPCode computes entry's current TOTP code at now (spec
// section 14.1: "see the current code").
func GenerateTOTPCode(entry domain.Entry, now time.Time) (string, error) {
	params, err := entryTOTPParams(entry)
	if err != nil {
		return "", err
	}
	return totp.Generate(params, now)
}

// TOTPTimeRemaining returns how much of the current TOTP period is left
// at now (spec section 14.1: "see the countdown of the current
// 30-second window") — the time until GenerateTOTPCode would start
// returning a different code.
func TOTPTimeRemaining(entry domain.Entry, now time.Time) (time.Duration, error) {
	params, err := entryTOTPParams(entry)
	if err != nil {
		return 0, err
	}

	period := params.Period
	if period <= 0 {
		period = 30 * time.Second
	}
	elapsed := time.Duration(now.Unix()%int64(period/time.Second)) * time.Second
	return period - elapsed, nil
}
