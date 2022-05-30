module github.com/icedream/livestream-tools/icedreammusic/foobar2000

go 1.16

replace (
	github.com/icedream/livestream-tools/icedreammusic/metacollector => ../metacollector
	github.com/icedream/livestream-tools/icedreammusic/tuna => ../tuna
)

require (
	github.com/billziss-gh/cgofuse v1.5.0
	github.com/dhowden/tag v0.0.0-20220530110423-77907a30b7f1
	github.com/gin-gonic/gin v1.7.7
	github.com/icedream/livestream-tools/icedreammusic/metacollector v0.0.0-00010101000000-000000000000
	github.com/icedream/livestream-tools/icedreammusic/tuna v0.0.0-00010101000000-000000000000
	github.com/karalabe/xgo v0.0.0-20191115072854-c5ccff8648a7 // indirect
	github.com/nicksnyder/go-i18n v1.10.1 // indirect
	gopkg.in/alecthomas/kingpin.v3-unstable v3.0.0-20191105091915-95d230a53780
)
