module okxbot

go 1.20

require (
	github.com/emersion/go-imap v1.2.1
	github.com/emersion/go-message v0.15.0
	github.com/joho/godotenv v1.5.1
	github.com/pquerna/otp v1.4.0
	structs v0.0.0
)

require (
	github.com/boombuler/barcode v1.0.1-0.20190219062509-6c824513bacc // indirect
	github.com/emersion/go-sasl v0.0.0-20200509203442-7bfe0ed36a21 // indirect
	github.com/emersion/go-textwrapper v0.0.0-20200911093747-65d896831594 // indirect
	golang.org/x/text v0.3.7 // indirect
)

replace structs => ./structs
