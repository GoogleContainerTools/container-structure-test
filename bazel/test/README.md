# Bazel smoke test

Verifies that the container_structure_test bazel rule exposed by this project works properly.

```sh
cd bazel/test
bazel test ...
bazel test --enable_bzlmod ...
```