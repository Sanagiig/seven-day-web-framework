module seven-day-web-framework

require (
	geeCache v0.0.0
	github.com/go-sql-driver/mysql v1.8.1
	google.golang.org/protobuf v1.33.0
)

require (
	common v0.0.0 // indirect
	filippo.io/edwards25519 v1.1.0 // indirect
)

replace (
	common => ./common
	gee => ./gee
	geeCache => ./geeCache
)

go 1.22.0
