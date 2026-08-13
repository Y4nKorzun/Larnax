package application

// Version is this application's version — spec section 22.2's :doctor
// "Application: 0.1.0" line, and spec section 22.1's allowed-to-log
// application version — kept as the one place callers like a future
// cmd/kdbx-tui/main.go and BuildDoctorReport's own callers get it from,
// rather than each hardcoding the string separately.
//
// 0.1.0 marks this as pre-release, tracked against spec section 26's P0
// acceptance criteria rather than semantic API stability.
const Version = "0.1.0"
