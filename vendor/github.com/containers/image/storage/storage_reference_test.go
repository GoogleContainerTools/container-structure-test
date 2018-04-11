// +build !containers_image_storage_stub

package storage

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStorageReferenceTransport(t *testing.T) {
	newStore(t)
	ref, err := Transport.ParseReference("busybox")
	require.NoError(t, err)
	transport := ref.Transport()
	st, ok := transport.(*storageTransport)
	require.True(t, ok)
	assert.Equal(t, *(Transport.(*storageTransport)), *st)
}

func TestStorageReferenceDockerReference(t *testing.T) {
	newStore(t)
	ref, err := Transport.ParseReference("busybox")
	require.NoError(t, err)
	dr := ref.DockerReference()
	require.NotNil(t, dr)
	assert.Equal(t, "docker.io/library/busybox:latest", dr.String())

	ref, err = Transport.ParseReference("@" + sha256digestHex)
	require.NoError(t, err)

	dr = ref.DockerReference()
	assert.Nil(t, dr)
}

// A common list of reference formats to test for the various ImageReference methods.
var validReferenceTestCases = []struct {
	input, canonical string
	namespaces       []string
}{
	{
		"busybox", "docker.io/library/busybox:latest",
		[]string{"docker.io/library/busybox", "docker.io/library", "docker.io"},
	},
	{
		"example.com/myns/ns2/busybox:notlatest", "example.com/myns/ns2/busybox:notlatest",
		[]string{"example.com/myns/ns2/busybox", "example.com/myns/ns2", "example.com/myns", "example.com"},
	},
	{
		"@" + sha256digestHex, "@" + sha256digestHex,
		[]string{},
	},
	{
		"busybox@" + sha256digestHex, "docker.io/library/busybox:latest@" + sha256digestHex,
		[]string{"docker.io/library/busybox", "docker.io/library", "docker.io"},
	},
	{
		"busybox@sha256:" + sha256digestHex, "docker.io/library/busybox@sha256:" + sha256digestHex,
		[]string{"docker.io/library/busybox", "docker.io/library", "docker.io"},
	},
}

func TestStorageReferenceStringWithinTransport(t *testing.T) {
	store := newStore(t)
	optionsList := ""
	options := store.GraphOptions()
	if len(options) > 0 {
		optionsList = ":" + strings.Join(options, ",")
	}
	storeSpec := fmt.Sprintf("[%s@%s+%s%s]", store.GraphDriverName(), store.GraphRoot(), store.RunRoot(), optionsList)

	for _, c := range validReferenceTestCases {
		ref, err := Transport.ParseReference(c.input)
		require.NoError(t, err, c.input)
		assert.Equal(t, storeSpec+c.canonical, ref.StringWithinTransport(), c.input)
	}
}

func TestStorageReferencePolicyConfigurationIdentity(t *testing.T) {
	store := newStore(t)
	storeSpec := fmt.Sprintf("[%s@%s]", store.GraphDriverName(), store.GraphRoot())

	for _, c := range validReferenceTestCases {
		ref, err := Transport.ParseReference(c.input)
		require.NoError(t, err, c.input)
		assert.Equal(t, storeSpec+c.canonical, ref.PolicyConfigurationIdentity(), c.input)
	}
}

func TestStorageReferencePolicyConfigurationNamespaces(t *testing.T) {
	store := newStore(t)
	storeSpec := fmt.Sprintf("[%s@%s]", store.GraphDriverName(), store.GraphRoot())

	for _, c := range validReferenceTestCases {
		ref, err := Transport.ParseReference(c.input)
		require.NoError(t, err, c.input)
		expectedNS := []string{}
		for _, ns := range c.namespaces {
			expectedNS = append(expectedNS, storeSpec+ns)
		}
		expectedNS = append(expectedNS, storeSpec)
		expectedNS = append(expectedNS, fmt.Sprintf("[%s]", store.GraphRoot()))
		assert.Equal(t, expectedNS, ref.PolicyConfigurationNamespaces())
	}
}

// NewImage, NewImageSource, NewImageDestination, DeleteImage tested in storage_test.go
