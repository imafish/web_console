load("@io_bazel_rules_go//go:def.bzl", "go_binary")

go_binary(
  name = "server",
  srcs = ["server.go"],
  goarch = "amd64",
  goos = "linux",
  deps = ["//api/proto:api_grpc"],
)

go_binary(
  name = "server_macos_arm64",
  srcs = ["server.go"],
  goarch = "arm64",
  goos = "darwin",
  deps = ["//api/proto:api_grpc"],
)
