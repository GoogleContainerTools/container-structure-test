py_library(
    name = "example_lib",
    srcs = glob([
        "*.py",
    ]),
    deps = ["@containerregistry"],
)

py_test(
    name = "example_test",
    srcs = ["example_test.py"],
    deps = [
        ":example_lib",
        "//testing/lib:containerregistry_mock_lib",
        "@containerregistry",
    ],
)
