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
      "//third-party:com_github_bazelbuild_bazel_gazelle@v0.35.0",
      "//third-party:com_github_bazelbuild_bazel_gazelle_merger@v0.35.0",
      "//third-party:com_github_bazelbuild_bazel_gazelle_rule@v0.35.0",
      "//third-party:com_github_bazelbuild_buildtools@v0.0.0-20240207142252-03bf520394af",
      "//third-party:com_github_bazelbuild_buildtools_build@v0.0.0-20240207142252-03bf520394af",
      ":geomys_lib",
    ],
)

go_library(
    name = "geomys_lib",
    srcs = ["deps.go", "canonicalize.go"],
    package_name = "github.com/dvtkrlbs/geomys"
#    deps = [
#      "//third-party:com_github_bazelbuild_bazel_gazelle",
#      "//third-party:com_github_bazelbuild_buildtools",
#    ],
)