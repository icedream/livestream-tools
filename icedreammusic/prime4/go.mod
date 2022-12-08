module github.com/icedream/livestream-tools/icedreammusic/prime4

go 1.19

replace (
	github.com/icedream/livestream-tools/icedreammusic/metacollector => ../metacollector
	github.com/icedream/livestream-tools/icedreammusic/tuna => ../tuna
)

require (
	github.com/icedream/go-stagelinq v0.0.1
	github.com/icedream/livestream-tools/icedreammusic/metacollector e27172751086
	github.com/icedream/livestream-tools/icedreammusic/tuna d83cb4af0567
)

require golang.org/x/text v0.4.0 // indirect
