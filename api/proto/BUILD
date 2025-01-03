load("@rules_proto//proto:defs.bzl", "proto_library")
load("@rules_proto_grpc//grpc:defs.bzl", "grpc_library")

proto_library(
  name = "api_proto",
  srcs = ["api.proto"],
  visibility = ["//visibility:public"],
)

grpc_library(
  name = "api_grpc",
  srcs = [":api_proto"],
  visibility = ["//visibility:public"],
  deps = [":api_proto"],
  output_dir = "../../",
)
