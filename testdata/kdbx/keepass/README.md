## Fixture provenance

`kdbx4-example.kdbx` is `tests/kdbx4/example.kdbx` from
[tobischo/gokeepasslib](https://github.com/tobischo/gokeepasslib) at tag
`v3.7.0`, redistributed under that project's MIT license
(Copyright (c) 2024 Tobias Schoknecht).

It is included here specifically because it is genuinely KeePass-generated,
not gokeepasslib self-round-tripped: `kdbx41_test.go` in that project
describes its own `tests/kdbx41/example.kdbx` fixture as having been "created
out of the **KeePass generated** tests/kdbx4/example.kdbx". Testing our own
decode path against it validates behavior against a real independent writer,
which a library round-tripping its own output cannot do (see gokeepasslib
issue #150 for exactly the failure mode that distinction catches).

Password: `abcdefg12345678`

Known content (from gokeepasslib's own `decoder_test.go`), used as this
repo's assertions in `internal/infrastructure/kdbx/fixture_test.go`:
- `Root.Groups[0].Groups[0].Entries[0]` password: `Password`
- `Root.Groups[0].Groups[0].Entries[1]` password: `AnotherPassword`
- `Root.Groups[0].Groups[1].Entries[0]` has one binary attachment containing
  the text `Hello world`

Contains only throwaway test credentials — nothing sensitive.
