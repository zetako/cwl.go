module github.com/lijiang2014/cwl.go

go 1.12

require (
	github.com/google/uuid v1.3.0
	github.com/grpc-ecosystem/go-grpc-middleware v1.4.0
	github.com/otiai10/mint v1.3.2
	github.com/robertkrimen/otto v0.0.0-20211008084715-4eacda02dd21
	github.com/spf13/cast v1.5.0
	github.com/spf13/cobra v1.7.0
	go.uber.org/zap v1.19.0
	google.golang.org/grpc v1.55.0
	google.golang.org/protobuf v1.30.0
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.1
	starlight v0.0.0-00010101000000-000000000000
)

require (
	github.com/dgrijalva/jwt-go v3.2.0+incompatible // indirect
	github.com/golang/mock v1.6.0 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/jinzhu/gorm v1.9.16 // indirect
	github.com/olivere/elastic/v7 v7.0.32 // indirect
	github.com/sirupsen/logrus v1.7.0 // indirect
	github.com/zetako/scontrol v1.0.0
)

replace github.com/lijiang2014/cwl.go => ./

replace starlight => ../starlight
