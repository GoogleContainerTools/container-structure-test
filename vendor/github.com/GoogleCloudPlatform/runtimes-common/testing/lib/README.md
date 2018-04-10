## containerregistry Testing Lib

This is a Python library for testing [containerregistry](https://github.com/google/containerregistry). It creates a mock registry to store images in. It also mocks docker\_session.Push and docker\_image.FromRegistry, which interact with the mock registry.

## Usage 

To set up the testing library, the test class should inherit from MockRegistryTestBase, and should call its setUp function. This creates the mock registry as self.registry.  Any calls to session.Push and docker_image.FromRegistry will now interact with this mock registry.

```python
from testing.lib import mock_registry_test_base
import unittest

class SampleTest(mock_registry_test_base.MockRegistryTestBase):
    def setUp(self):
        super(SampleTest, self).setUp()
    
    # Tests ...

if __name__ == '__main__':
    unittest.main()
```

Be sure to add the testing library as a dependency of your test in your BUILD file.

## API

The mock registry can now be set up as desired with the following functions:

```python
# setTags associates the image name full_repo with the specified tags
setTags(full_repo, tags)
    # full_repo: image name with tag or digest
    # tags: array of associated tags

# setImage associates the image name full_repo with an image
setImage(full_repo, image)
    # full_repo: image name with tag or digest
    # image: image to associate the name with

# setManifests associates the repository with a manifest
setManifests(repository, manifest)
    # repository: image name
    # manifest: manifest to associate image name with

# clearRegistry clears the mock registry
clearRegistry()
```

The following FromRegistry functions will interact with the mock registry when called:

```python
exists()
manifests()
tags()
```

The following session.Push functions will interact with the mock registry when called:
```python
upload(src_image, use_digest=False)
```

## Registry Assertions
To assert the status of an image in the registry, you can use the following:

```python
# AssertPushed asserts the image was pushed to the registry
AssertPushed(registry, image)
    # registry: of type MockRegistry
    # image: name of the image

# AssertNotPushed asserts the image was not pushed to the registry
AssertNotPushed(registry, image)
    # registry: of type MockRegistry
    # image: name of the image
```

## Example

The following sample code uses the containerregistry library and has an associated test which uses this testing library. This code and the associated BUILD file can be found in the example directory.

##### example.py
```python
from containerregistry.client.v2_2 import docker_image
from containerregistry.client.v2_2 import docker_session
from containerregistry.client import docker_creds
from containerregistry.client import docker_name
from containerregistry.transport import transport_pool
import httplib2

DIGEST = '0000000000000000000000000000000000000000000000000000000000000000'

def main():
    digest = 'fake.gcr.io/test/test@sha256:' + DIGEST
    tag = 'fake.gcr.io/test/test:tag'
    src_name = docker_name.Digest(digest)
    dest_name = docker_name.Tag(tag)
    creds = docker_creds.DefaultKeychain.Resolve(src_name)
    transport = transport_pool.Http(httplib2.Http)

    with docker_image.FromRegistry(src_name, creds, transport) as src_img:
        if src_img.exists():
            creds = docker_creds.DefaultKeychain.Resolve(dest_name)
            with docker_session.Push(dest_name, creds, transport) as push:
                    push.upload(src_img)

if __name__ == '__main__':
    main()

```

##### example_test.py
```python
from testing.lib import mock_registry_test_base
from containerregistry.client.v2_2 import docker_image
import example
import unittest

DIGEST = '0000000000000000000000000000000000000000000000000000000000000000'

class ExampleTest(mock_registry_test_base.MockRegistryTestBase):

    def setUp(self):
        super(ExampleTest, self).setUp()

    def testMain(self):
        # Add initial image to registry
        with docker_image.FromRegistry('fake.gcr.io/test/test@sha256:' + DIGEST) as img:
            self.registry.setImage('fake.gcr.io/test/test@sha256:' + DIGEST, img)

        example.main()
        # Assert that new image was pushed correctly
        self.AssertPushed(self.registry, "fake.gcr.io/test/test:tag")

if __name__ == '__main__':
    unittest.main()

```



