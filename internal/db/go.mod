module db

go 1.23.3

replace internal/pb => ../pb

require (
	github.com/mattn/go-sqlite3 v1.14.16
	google.golang.org/protobuf v1.35.2
	internal/pb v1.0.0
)

require (
	golang.org/x/net v0.29.0 // indirect
	golang.org/x/sys v0.25.0 // indirect
	golang.org/x/text v0.18.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240903143218-8af14fe29dc1 // indirect
	google.golang.org/grpc v1.68.1 // indirect
)
