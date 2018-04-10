## Retagger

The retagger is a tool to manage the tags on container images in GCR. It can be run to update tags on container images or to verify that expected tags are the same as in production.

To use the retagger, you will need a config file in [this](https://github.com/GoogleCloudPlatform/runtimes-common/blob/master/reconciletags/sample.json) format. This JSON tells the retagger which container images need to be retagged, and what tags should be associated with each of them.

## Build

The retagger can be built as an executable par file with Bazel, by running the following command from the root directory:

```
bazel build reconciletags:reconciletags.par
```

To run the retagger:

```
reconciletags.par /path/to/config/file
```

To verify the config file is valid, use the --dry-run flag:

```
reconciletags.par --dry-run /path/to/config/file
```

To verify the specified tags in the config file are the same as in production, use the --data-integrity flag:
```
reconciletags.par --data-integrity /path/to/config/file
```
