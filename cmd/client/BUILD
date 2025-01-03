load("@io_bazel_rules_go//go:def.bzl", "go_binary")

go_binary(
  name = "client",
  srcs = ["main.go"],
  goarch = "amd64",
  goos = "linux",
  deps = ["//api/proto:api_grpc"],
)

go_binary(
  name = "client_arm64_macos",
  srcs = ["main.go"],
  goarch = "arm64",
  goos = "darwin",
  deps = ["//api/proto:api_grpc"],
)
