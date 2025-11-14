# Bazel smoke test

Verifies that the container_structure_test bazel rule exposed by this project works properly.

## Running tests with pre-compiled toolchain

```sh
cd bazel/test
bazel test ...
bazel test --enable_bzlmod ...
```

## Testing with local changes (non-pre-compiled toolchain)

When developing changes, you may want to test your modifications before they're compiled a release. Here's how to test with a locally built binary:

1. Build your local binary:
   ```sh
   go build -o /tmp/container-structure-test-local ./cmd/container-structure-test/
   ```

2. Temporarily modify `bazel/container_structure_test.bzl`:
   ```sh
   sed -i.bak 's|readonly st=$(rlocation {st_path})|readonly st="/tmp/container-structure-test-local"|g' bazel/container_structure_test.bzl
   ```

3. Run the bazel test:
   ```sh
   cd bazel/test
   bazel test :test --test_output=all
   ```

4. Restore the original rule:
   ```sh
   mv bazel/container_structure_test.bzl.bak bazel/container_structure_test.bzl
   ```

This allows you to verify that your changes work correctly with the bazel integration before submitting them.
