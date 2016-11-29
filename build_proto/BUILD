load("@io_bazel_rules_go//proto:go_proto_library.bzl", "go_proto_library")

genrule(
    name = "copy_and_fix",
    srcs = ["@io_bazel//src/main/protobuf:srcs"],
    outs = ["build.proto"],
    cmd = ("SRCS=($(locations @io_bazel//src/main/protobuf:srcs));" +
           "cp $$(dirname $$SRCS)/build.proto $@"),
)

go_proto_library(
    name = "go_default_library",
    srcs = ["build.proto"],
    visibility = ["//visibility:public"],
)
