#!/usr/bin/python

CLUSTER_NAME = 'gcp-integration-test-cluster'

LOGNAME_LENGTH = 16

LOGGING_PREFIX = 'GCP_INTEGRATION_TEST_'

DEFAULT_TIMEOUT = 30  # seconds

ROOT_ENDPOINT = '/'
ROOT_EXPECTED_OUTPUT = 'Hello World!'

STANDARD_LOGGING_ENDPOINT = '/logging_standard'
CUSTOM_LOGGING_ENDPOINT = '/logging_custom'
MONITORING_ENDPOINT = '/monitoring'
EXCEPTION_ENDPOINT = '/exception'
CUSTOM_ENDPOINT = '/custom'
ENVIRONMENT_ENDPOINT = '/environment'

METRIC_PREFIX = 'custom.googleapis.com/{0}'
METRIC_TIMEOUT = 60  # seconds

# subset of levels found at
# https://cloud.google.com/logging/docs/reference/v2/rest/v2/LogEntry#logseverity
SEVERITIES = [
    'WARNING',
    'ERROR',
    'CRITICAL'
]

GAE = 'GAE'
GKE = 'GKE'
ENVS = [GAE, GKE]
