module github.com/icedream/livestream-tools/icedreammusic/prime4

go 1.21.0

toolchain go1.24.2

replace (
	github.com/icedream/livestream-tools/icedreammusic/metacollector => ../metacollector
	github.com/icedream/livestream-tools/icedreammusic/tuna => ../tuna
)

require (
	github.com/icedream/go-stagelinq v1.0.0
	github.com/icedream/livestream-tools/icedreammusic/metacollector v0.0.0-20240122014424-f96ec7e413e0
	github.com/icedream/livestream-tools/icedreammusic/tuna v0.0.0-20250428161951-7b4908eb159b
)

require golang.org/x/text v0.21.0 // indirect
