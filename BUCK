load("@prelude//toolchains:go.bzl", "system_go_toolchain")


go_binary(
    name = "example",
    srcs=[
      "cmd/hello/main.go",
    ],
    deps = [
      "//third-party:com_github_burntsushi_toml@v0.3.1"
    ],
)


go_binary(
    name = "geomys",
    srcs=[
      "cmd/geomys/main.go",
    ],
    deps = [
      "//third-party:com_github_bazelbuild_bazel_gazelle",
      "//third-party:com_github_bazelbuild_buildtools",
      ":geomys_lib",
    ],
)

go_library(
    name = "geomys_lib",
    srcs = glob(["deps.go, canonicalize.go"]),
    deps = [
#      "//third-party:com_github_bazelbuild_bazel_gazelle",
#      "//third-party:com_github_bazelbuild_buildtools",
    ],
)