module github.com/StollD/proton-drive

go 1.21

toolchain go1.22.2

require (
	github.com/ProtonMail/gopenpgp/v2 v2.7.5
	github.com/barweiss/go-tuple v1.1.2
	github.com/deckarep/golang-set/v2 v2.6.0
	github.com/henrybear327/go-proton-api v1.0.0
	github.com/relvacode/iso8601 v1.4.0
	golang.org/x/time v0.5.0
)

require (
	github.com/ProtonMail/bcrypt v0.0.0-20211005172633-e235017c1baf // indirect
	github.com/ProtonMail/gluon v0.17.1-0.20240423123310-0266b0f75d41 // indirect
	github.com/ProtonMail/go-crypto v1.0.0 // indirect
	github.com/ProtonMail/go-mime v0.0.0-20230322103455-7d82a3887f2f // indirect
	github.com/ProtonMail/go-srp v0.0.7 // indirect
	github.com/PuerkitoBio/goquery v1.9.2 // indirect
	github.com/andybalholm/cascadia v1.3.2 // indirect
	github.com/bradenaw/juniper v0.15.3 // indirect
	github.com/cloudflare/circl v1.3.8 // indirect
	github.com/cronokirby/saferith v0.33.0 // indirect
	github.com/emersion/go-message v0.18.1 // indirect
	github.com/emersion/go-vcard v0.0.0-20230815062825-8fda7d206ec9 // indirect
	github.com/go-resty/resty/v2 v2.12.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	golang.org/x/crypto v0.22.0 // indirect
	golang.org/x/exp v0.0.0-20240416160154-fe59bbe5cc7f // indirect
	golang.org/x/net v0.24.0 // indirect
	golang.org/x/sync v0.7.0 // indirect
	golang.org/x/sys v0.19.0 // indirect
	golang.org/x/text v0.14.0 // indirect
)

replace github.com/henrybear327/go-proton-api v1.0.0 => github.com/StollD/go-proton-api v0.0.0-20240501114039-b4b2f7d99b66
