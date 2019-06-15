module github.com/PenF00k/go-wand

go 1.12

require (
	github.com/fsnotify/fsnotify v1.4.7
	github.com/golang/protobuf v1.3.1
	github.com/gorilla/websocket v1.4.0
	github.com/iancoleman/strcase v0.0.0-20190422225806-e506e3ef7365
	github.com/jessevdk/go-assets v0.0.0-20160921144138-4f4301a06e15
	github.com/satori/go.uuid v1.2.0
	github.com/sirupsen/logrus v1.4.2
	github.com/urfave/cli v1.20.0
	gitlab.vmassive.ru/wand v0.0.0
	google.golang.org/grpc v1.21.1
	gopkg.in/yaml.v2 v2.2.2
)

replace gitlab.vmassive.ru/wand => ./
