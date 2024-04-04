module seven-day-web-framework

require geeCache v0.0.0

require (
	common v0.0.0 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	google.golang.org/protobuf v1.33.0 // indirect
)

replace (
	common => ./common
	gee => ./gee
	geeCache => ./geeCache
)

go 1.22
