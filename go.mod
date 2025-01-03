module example.com/web_console

go 1.23.3

replace example.com/web_console/internal/pb => ./internal/pb

replace example.com/web_console/internal/db => ./internal/db

require (
	example.com/web_console/internal/pb v1.0.0
	example.com/web_console/internal/db v1.0.0
	google.golang.org/grpc v1.68.1
)

require (
	golang.org/x/net v0.29.0 // indirect
	golang.org/x/sys v0.25.0 // indirect
	golang.org/x/text v0.18.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240903143218-8af14fe29dc1 // indirect
	google.golang.org/protobuf v1.35.2 // indirect
)
