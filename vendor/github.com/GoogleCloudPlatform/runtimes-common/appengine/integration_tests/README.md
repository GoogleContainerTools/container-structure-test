Integration Testing Framework
=============

This code builds an image which serves as a framework to run basic integration tests for a language runtime. It relies on the existence of a sample application which fulfills a spec (detailed below), that will be built and deployed on top of either a newly built runtime image, or built on the language runtime through the gcloud builder flow, using a newly built runtime builder image. These tests are intended to be run as part of a post-push verification step in a continuous deployment system. The test driver can be run directly via any build host, or as a build step in a [Google Cloud Container Build](https://cloud.google.com/container-builder/docs/overview)).

To run these tests through a cloudbuild, add the following build step to the **end** of your build config (cloudbuild.yaml or cloudbuild.json):

	name: gcr.io/gcp-runtimes/integration_test
	args: [
		'-i', <target_staging_image>,
		'-d', <sample_application_directory>
	]

**It's crucial that this step appears after your image under test has been built**; without a built image, there will be nothing to test, and your build will fail!


The sample application directory should contain the application fulfilling the integration test spec, as well as the necessary build artifacts to deploy the application via gcloud, which means at minimum:

* a templated `Dockerfile.in`, with the first line being
	` FROM ${STAGING_IMAGE} `

	**OR**

* a templated `test.yaml.in`, containing a step for the runtime builder that looks like:

	- name: ${STAGING_BUILDER_IMAGE} 
	- args: [] (these are optional)

	**AND**

* an `app.yaml` config

If the tests are being deployed using a staging runtime builder image, the following gcloud config values **must be set, or else the build will fail**:

* `app/use_runtime_builders` set to `True` (tells the gcloud installation to use the runtime builder flow)
* `app/runtime_builders_root` set to *the location of the runtime manifest* (tells gcloud where to look for your local runtime builder, which in turn contains the staging runtime builder image).

Alternatively, the application can be manually deployed *before* running the tests; in this scenario, the '--no-deploy' flag can be passed to the build step to opt out of deploying, in tandem with the URL at which the deployed application can be accessed:

	args: [
		...,
		'--no-deploy',
		'--url', '<application_url>'
	]

In addition, each test (detailed below) can be skipped by passing the corresponding flag to the build step. For example, to opt out of the monitoring test in the run, pass the `--skip-monitoring-tests` flag to the build step.

## Tests

### Serving (Root)
#####` - GET http://<application_url>/`

*Response*

- If successful, the application should respond with the string 'Hello World!'

This is the most basic integration test. The driver performs a GET request to the root endpoint at the deployed application's URL. It retrieves data from the application, and verifies it matches the expected output (the text ‘Hello World!’).


### Standard Logging
#####` - POST http://<application_url>/logging_standard`

*Request Body*

| Property Name | Value | Description |
| --- | --- | --- |
| token | 16 character (8 byte) hexadecimal string (uppercase) | The token to write in Stackdriver logs. |
| level | string; alphabetic (uppercase) | String representing the severity of the log entry. |

*Response*

- If successful, the application should respond with the source that the logs were written to in Stackdriver, and a 200 response code.

This tests the runtime’s integration with Stackdriver Logging through the runtime's standard logging module. The driver will generate a token and a log level, and send this payload to the sample application via a POST request. Once the application receives the payload, it will log the token at the provided level through its standard logging module (e.g. [logging](https://docs.python.org/2/library/logging.html) in Python). The application will then send back to the test driver the location in Stackdriver to which the log entry was written; this varies across runtimes. See the [Logging v2 API Docs](https://cloud.google.com/logging/docs/reference/v2/rest/v2/LogEntry) for details; as an example, integration with Python's standard logging module sends all logs to `projects/<project_id>/logs/appengine.googleapis.com%2Fstderr`, so the sample application will return back to the test driver `appengine.googleapis.com%2Fstderr`.

If the write succeeded, the test driver will wait a short period of time for the log entry to propagate through Stackdriver, then retrieve all log entries written to the provided log source in the previous 2 minutes and verify that the provided token/level pair appears in Stackdriver.


### Custom Logging
#####` - POST http://<application_url>/logging_custom`

*Request Body*

| Property Name | Value | Description |
| --- | --- | --- |
| log_name      | string; 16 characters; alphabetic (upper/lower case) | The name of the log to write in Stackdriver. |
| token | 16 character (8 byte) hexadecimal string (uppercase) | The token to write in Stackdriver logs. |
| level | string; alphabetic (uppercase) | String representing the severity of the log entry. |

*Response*

- If successful, the application should respond with the source that the logs were written to in Stackdriver, and a 200 response code.

This tests the runtime’s integration with writing custom log entries to Stackdriver Logging through a client library, instead of using the runtime's standard logging module's integration with the Stackdriver logging API. The driver will generate a payload containing a log name, token, and level, and send this payload to the sample application via a POST request. Once the application receives the payload, it will write a log entry to Stackdriver with the provided name and token value, at the specified level. *This is commonly done by using a language-specific client library, such as [google-cloud-python](https://github.com/GoogleCloudPlatform/google-cloud-python), though the implementation is left up to the runtime maintainers.* The application will then signal back to the driver that either the log entry has been written successfully by providing it with the source to which the log entry was written, or that an error was encountered (causing the test to fail). If the write succeeded, the test driver will wait a short period of time for the log entry to propagate through Stackdriver, then retrieve all log entries written to that specified log name in the previous 2 minutes and verify that the provided token/level pair appears in Stackdriver.


### Monitoring
#####` - POST http://<application_url>/monitoring`

*Request Body*

| Property Name | Value | Description |
| --- | --- | --- |
| name      | string; 16 characters; alphabetic (upper/lower case) | The name of the metric to write in Stackdriver; **always prefixed with ‘custom.googleapis.com/’**. |
| token | integer 64-bit | The metric value to write into Stackdriver. |

*Response*

- If successful, the application should respond with a 200 response code and the string ‘OK’.

The monitoring test confirms the runtime’s ability to write metrics to Stackdriver. The driver will first generate a request payload and send it to the sample application in the normal fashion (HTTP POST request). Once the application receives the payload, it will create a [Cloud Monitoring Client](https://github.com/GoogleCloudPlatform/google-cloud-python/blob/master/monitoring/google/cloud/monitoring/client.py), and use this client to either retrieve the metric descriptor corresponding to the provided name, or create a new metric descriptor with this name (see the [Stackdriver Custom Metric Documentation](https://cloud.google.com/monitoring/custom-metrics/creating-metrics) for more information on this). Once the metric descriptor is retrieved, the client will use this descriptor and the provided payload to write a custom metric to Stackdriver, and signal back to the driver that either the metric was written successfully, or that an error was encountered (causing the test to fail). If the write succeeded, the test driver will wait a short period of time for the metric entry to propagate through Stackdriver, then retrieve all metric entries with the specified name, and verify that the provided token exists as one of the values in that query.


### Error Reporting/Exceptions
#####` - POST http://<application_url>/exception`

*Request Body*

| Property Name | Value | Description |
| --- | --- | --- |
| token | integer 64-bit | The metric value to write into Stackdriver. |

*Response*

- If successful, the application should respond with a 200 response code and the string ‘OK’.

This tests the runtime’s ability to report errors to Stackdriver. The driver will generate a request payload and POST it to the sample application. Upon receiving the payload, the application will create a Cloud Error Reporting Client. It will then use this client to report a generic exception. Additionally, it will use the client to report an exception with the provided token. Finally, it will report back to the test driver indicating that the exceptions were recorded successfully, or that an error was encountered, failing the test.

At this time, there is no support within the google-cloud-python client library for reading exceptions from Stackdriver, though the public API exists. It is potentially planned to add support within this client library to do this, enabling the test driver to actually verify the specific exception was recorded in Stackdriver successfully.


### Custom Tests
#####` - GET http://<application_url>/custom`

The integration test framework also contains support for running custom tests inside of the sample application, and reporting the results of these tests through the integration framework's report. 
This provides a convenient way of running standard and custom tests inside of the same test run.
The test driver will make a GET request to `/custom` at the sample application's URL, and retrieve a list of test configuration specs (see below) that specify the custom tests the sample application is set up to run. 
The driver will then enter two phases for each test:
* The Execution: the driver will make sequential requests to each of the specified paths and store the results.
* The Validation: the results of the executions will be asserted using the provided specification.

One final requirement for the test driver is that it exposes the cloud environment where it is deployed to the test driver. This allows the driver to dynamically test different scenarios depending on the environment (e.g. log location). **The test application must expose its cloud environment at http://<application_url>/environment!** Tests will not function properly otherwise.

At this URL, the application should return the cloud environment where it itself is deployed, in the form of a string. Valid return values in the response are:

* `'GAE'`
* `'GKE'`

There are a few ways to retrieve this information from within the deployed application. A generally accepted way is to check for GAE-specific environment variables on the host; these are always set by GAE upon deploy, so if they don't exist we know we're not on GAE, and can assume we're deployed in a GKE cluster.

For example, the implementation in the python application:

```
_APPENGINE_FLEXIBLE_ENV_VM = 'GAE_APPENGINE_HOSTNAME'
"""Environment variable set in App Engine when vm:true is set."""

_APPENGINE_FLEXIBLE_ENV_FLEX = 'GAE_INSTANCE'
"""Environment variable set in App Engine when env:flex is set."""

...

return (_APPENGINE_FLEXIBLE_ENV_VM in os.environ or
        _APPENGINE_FLEXIBLE_ENV_FLEX in os.environ), 200
```


#### The Execution phase
The execution can be configured using the field `steps`. This field contains an array of `Step`, where each `Step` represent a request 
and an associated configuration, a `Step` is defined using the following schema:
  * `name`: (optional) the name of the step.
  * `path`: (required) the path at which the request will be send.
  * `configuration`: (optional) the configuration of the request.
    * `headers`: (optional) an object (key/value) representing the headers and their values.
    * `method`: (optional, default to `GET`): The method to use for the request (`GET`/`POST`).
    * `content`: (optional) the body of the request.
    
#### The Validation phase
The results of the executions can be verified using the `validation` field. 
The assertions are defined in the field `match` which specify an array of `key` and `pattern` (regular expressions) to be compared.

The `key` represent a path into the execution results: `{step-name}.response.{field}`. 
The value present at `key` will be extracted and compared with the `pattern`.
The following fields are present in the response:
* status: (integer) the status code of the response.
* headers: (object) key/value representing the header names and their values.
* content: (text/object) the body of the response (if the header `Content-Type` is present and contains `application/json` 
the `content` is interpreted as a json object otherwise as plain text).

You can also use an `endpoint` in the test application to assert the result of the test. A request (POST) will be sent to
the specified endpoint, containing the context as a json payload, and the status code of the response will be asserted 
(2xx will considered as a success).

The schema for `validation` is the following:
* `validation`: the validation configuration
  * `match` (optional): an array of the expressions to be compared.
    * `key`: a path into the results of the steps.
    * `pattern`: a regular expression to be matched against the value at the path `key`.
  * `endpoint` (optional): A path in the test application to assert the result of the test.
    
These successes/failures will then be added to the integration test framework's report.

#### Example: Gzip headers

This example configuration tests the presence of the gzip headers. The request sent by the driver
to the endpoint `/gzip` will contains the header `Accept-Encoding` with the value `gzip`.
The validation phase will verify that the response's header `Content-Encoding` contains the value `gzip`.

```json
[{
  "name": "test-gzip",
  "steps": [{
    "name": "test-gzip-header",
    "configuration": {
      "headers": { "Accept-Encoding": "gzip" },
      "method": "GET"
    },
    "path": "/gzip"
  }],
  "validation": {
    "match": [{
      "key": "test-gzip-header.response.headers.Content-Encoding",
      "pattern": "gzip"
    }]
  },
  "timeout": 500
}]
```
