% registries.conf(5) System-wide registry configuration file
% Brent Baude
% Aug 2017

# NAME
registries.conf - Syntax of System Registry Configuration File

# DESCRIPTION
The REGISTRIES configuration file is a system-wide configuration file for container image
registries. The file format is TOML.

# FORMAT
The TOML_format is used to build simple list format for registries under two
categories: `search` and `insecure`. You can list multiple registries using
as a comma separated list.

Search registries are used when the caller of a container runtime does not fully specify the
container image that they want to execute.  These registries are prepended onto the front
 of the specified container image until the named image is found at a registry.

Insecure Registries.  By default container runtimes use TLS when retrieving images
from a registry.  If the registry is not setup with TLS, then the container runtime
will fail to pull images from the registry. If you add the registry to the list of
insecure registries then the container runtime will attempt use standard web protocols to
pull the image.  It also allows you to pull from a registry with self-signed certificates.
Note insecure registries can be used for any registry, not just the
registries listed under search.

The following example configuration defines two searchable registries and one
insecure registry.

```
[registries.search]
registries = ["registry1.com", "registry2.com"]

[registries.insecure]
registries = ["registry3.com"]
```

# HISTORY
Aug 2017, Originally compiled by Brent Baude <bbaude@redhat.com>
