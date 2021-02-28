module github.com/icedream/livestream-tools/icedreammusic/foobar2000

go 1.16

replace (
	github.com/icedream/livestream-tools/icedreammusic/metacollector => ../metacollector
	github.com/icedream/livestream-tools/icedreammusic/tuna => ../tuna
)

require (
	github.com/billziss-gh/cgofuse v1.4.0
	github.com/dhowden/tag v0.0.0-20201120070457-d52dcb253c63
	github.com/gin-gonic/gin v1.6.3
	github.com/icedream/livestream-tools/icedreammusic/metacollector v0.0.0-00010101000000-000000000000
	github.com/icedream/livestream-tools/icedreammusic/tuna v0.0.0-00010101000000-000000000000
)
