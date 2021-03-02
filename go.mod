module github.com/kmgreen2/agglo

go 1.15

require (
	cloud.google.com/go/storage v1.14.0
	github.com/DATA-DOG/go-sqlmock v1.5.0
	github.com/Knetic/govaluate v2.3.0+incompatible
	github.com/aws/aws-lambda-go v1.22.0
	github.com/aws/aws-sdk-go v1.37.12
	github.com/fsouza/fake-gcs-server v1.22.5
	github.com/golang/mock v1.4.4
	github.com/golang/protobuf v1.4.3
	github.com/google/uuid v1.1.2
	github.com/grpc-ecosystem/go-grpc-middleware v1.0.0
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.2.0
	github.com/minio/minio-go/v7 v7.0.9
	github.com/pkg/errors v0.9.1
	github.com/spf13/viper v1.4.0
	github.com/stretchr/testify v1.6.1
	go.etcd.io/bbolt v1.3.5
	go.opentelemetry.io/otel v0.15.0
	go.opentelemetry.io/otel/exporters/stdout v0.15.0
	go.opentelemetry.io/otel/exporters/trace/zipkin v0.15.0
	go.opentelemetry.io/otel/sdk v0.15.0
	go.uber.org/zap v1.10.0
	golang.org/x/net v0.0.0-20201224014010-6772e930b67b
	google.golang.org/api v0.40.0
	google.golang.org/genproto v0.0.0-20210226172003-ab064af71705
	google.golang.org/grpc v1.35.0
	google.golang.org/protobuf v1.25.0
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gorm.io/driver/postgres v1.0.8
	gorm.io/gorm v1.20.12
)

replace github.com/fsouza/fake-gcs-server => github.com/kmgreen2/fake-gcs-server v1.22.6-0.20210302174812-4d574d29511e
