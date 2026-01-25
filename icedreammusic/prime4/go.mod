module github.com/icedream/livestream-tools/icedreammusic/prime4

go 1.21.0

toolchain go1.25.6

replace (
	github.com/icedream/livestream-tools/icedreammusic/metacollector => ../metacollector
	github.com/icedream/livestream-tools/icedreammusic/tuna => ../tuna
)

require (
	github.com/icedream/go-stagelinq v1.0.0
	github.com/icedream/livestream-tools/icedreammusic/metacollector v0.0.0-20240122014424-f96ec7e413e0
	github.com/icedream/livestream-tools/icedreammusic/tuna v0.0.0-20221205042012-d83cb4af0567
)

require golang.org/x/text v0.21.0 // indirect
