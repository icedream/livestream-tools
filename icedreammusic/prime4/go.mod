module github.com/icedream/livestream-tools/icedreammusic/prime4

go 1.19

replace (
	github.com/icedream/livestream-tools/icedreammusic/metacollector => ../metacollector
	github.com/icedream/livestream-tools/icedreammusic/tuna => ../tuna
)

require (
	github.com/icedream/go-stagelinq v0.0.1
	github.com/icedream/livestream-tools/icedreammusic/metacollector v0.0.0-20221208055945-e27172751086
	github.com/icedream/livestream-tools/icedreammusic/tuna v0.0.0-20221205042012-d83cb4af0567
)

require golang.org/x/text v0.13.0 // indirect
