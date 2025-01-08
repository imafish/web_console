load("@rules_go//proto:def.bzl", "go_grpc_library")
load("@rules_proto//proto:defs.bzl", "proto_library")

proto_library(
    name = "proto_source",
    srcs = ["commandline_task.proto"],
    visibility = ["//visibility:public"],
    deps = [
        #"io_bazel_rules_go//proto/wkt:all",
        #        "@com_google_protobuf//:timestamp_proto",
        #       "@rules_proto//proto:timestamp_proto",
    ],
)

go_grpc_library(
    name = "proto",
    importpath = "internal/pb/",
    proto = ":proto_source",
    visibility = ["//visibility:public"],
)
