module github.com/fzxiao233/Vtb_Record

go 1.13

require (
	cloud.google.com/go/logging v1.0.0
	github.com/bitly/go-simplejson v0.5.0
	github.com/fsnotify/fsnotify v1.4.7
	github.com/fzxiao233/Go-Emoji-Utils v0.0.0-20200305114615-005e99b02c2f
	github.com/go-redis/redis v6.15.8+incompatible
	github.com/gogf/greuse v1.1.0
	github.com/hashicorp/golang-lru v0.5.4
	github.com/knq/sdhook v0.0.0-20190801142816-0b7fa827d09a
	github.com/mitchellh/mapstructure v1.1.2
	github.com/onsi/ginkgo v1.12.0 // indirect
	github.com/onsi/gomega v1.9.0 // indirect
	github.com/orandin/lumberjackrus v1.0.1
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/rclone/rclone v1.52.2
	github.com/recursionpharma/stackrus v0.0.0-20171005194045-12348afda34c
	github.com/sirupsen/logrus v1.6.0
	github.com/spf13/viper v1.6.2
	github.com/tidwall/gjson v1.6.0
	github.com/tidwall/pretty v1.0.1 // indirect
	github.com/valyala/bytebufferpool v1.0.0
	go.uber.org/ratelimit v0.1.0
	golang.org/x/crypto v0.0.0-20200423211502-4bdfaf469ed5 // indirect
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0
	google.golang.org/protobuf v1.23.0 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.0.0 // indirect
)

replace github.com/rclone/rclone v1.52.2 => github.com/NyaMisty/rclone v1.52.2
replace github.com/smallnest/ringbuffer => ../../smallnest/ringbuffer