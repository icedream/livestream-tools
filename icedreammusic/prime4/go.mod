module github.com/icedream/livestream-tools/icedreammusic/prime4

go 1.19

replace (
	github.com/icedream/livestream-tools/icedreammusic/metacollector => ../metacollector
	github.com/icedream/livestream-tools/icedreammusic/tuna => ../tuna
)

require (
	github.com/icedream/go-stagelinq v0.0.1
	github.com/icedream/livestream-tools/icedreammusic/metacollector v0.0.0-00010101000000-000000000000
	github.com/icedream/livestream-tools/icedreammusic/tuna v0.0.0-00010101000000-000000000000
)

require golang.org/x/text v0.3.7 // indirect
