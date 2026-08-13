## Wordlist provenance

`eff_large_wordlist.txt` is the [EFF Long Wordlist](https://www.eff.org/files/2016/07/18/eff_large_wordlist.txt)
(7,776 words, indexed by 5-digit dice-roll sequence), fetched from
eff.org and unmodified. It is the exact source spec section 31 names
for passphrase generation (sections 7.4, 10.5) and gives an entropy
table for (section 7.4).

Licensed CC BY (Creative Commons Attribution) per eff.org's site-wide
copyright notice. Source: Electronic Frontier Foundation, <https://www.eff.org/dice>.

The file is embedded verbatim via `go:embed` and parsed at package
init by `wordlist.go`, which also verifies the parsed count is exactly
7776 with no duplicates (see `wordlist_test.go`).
