package signature

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSignDockerManifest(t *testing.T) {
	mech, err := newGPGSigningMechanismInDirectory(testGPGHomeDirectory)
	require.NoError(t, err)
	defer mech.Close()

	if err := mech.SupportsSigning(); err != nil {
		t.Skipf("Signing not supported: %v", err)
	}

	manifest, err := ioutil.ReadFile("fixtures/image.manifest.json")
	require.NoError(t, err)

	// Successful signing
	signature, err := SignDockerManifest(manifest, TestImageSignatureReference, mech, TestKeyFingerprint)
	require.NoError(t, err)

	verified, err := VerifyDockerManifestSignature(signature, manifest, TestImageSignatureReference, mech, TestKeyFingerprint)
	assert.NoError(t, err)
	assert.Equal(t, TestImageSignatureReference, verified.DockerReference)
	assert.Equal(t, TestImageManifestDigest, verified.DockerManifestDigest)

	// Error computing Docker manifest
	invalidManifest, err := ioutil.ReadFile("fixtures/v2s1-invalid-signatures.manifest.json")
	require.NoError(t, err)
	_, err = SignDockerManifest(invalidManifest, TestImageSignatureReference, mech, TestKeyFingerprint)
	assert.Error(t, err)

	// Error creating blob to sign
	_, err = SignDockerManifest(manifest, "", mech, TestKeyFingerprint)
	assert.Error(t, err)

	// Error signing
	_, err = SignDockerManifest(manifest, TestImageSignatureReference, mech, "this fingerprint doesn't exist")
	assert.Error(t, err)
}

func TestVerifyDockerManifestSignature(t *testing.T) {
	mech, err := newGPGSigningMechanismInDirectory(testGPGHomeDirectory)
	require.NoError(t, err)
	defer mech.Close()
	manifest, err := ioutil.ReadFile("fixtures/image.manifest.json")
	require.NoError(t, err)
	signature, err := ioutil.ReadFile("fixtures/image.signature")
	require.NoError(t, err)

	// Successful verification
	sig, err := VerifyDockerManifestSignature(signature, manifest, TestImageSignatureReference, mech, TestKeyFingerprint)
	require.NoError(t, err)
	assert.Equal(t, TestImageSignatureReference, sig.DockerReference)
	assert.Equal(t, TestImageManifestDigest, sig.DockerManifestDigest)

	// Verification using a different canonicalization of TestImageSignatureReference
	sig, err = VerifyDockerManifestSignature(signature, manifest, "docker.io/"+TestImageSignatureReference, mech, TestKeyFingerprint)
	require.NoError(t, err)
	assert.Equal(t, TestImageSignatureReference, sig.DockerReference)
	assert.Equal(t, TestImageManifestDigest, sig.DockerManifestDigest)

	// For extra paranoia, test that we return nil data on error.

	// Invalid docker reference on input
	sig, err = VerifyDockerManifestSignature(signature, manifest, "UPPERCASEISINVALID", mech, TestKeyFingerprint)
	assert.Error(t, err)
	assert.Nil(t, sig)

	// Error computing Docker manifest
	invalidManifest, err := ioutil.ReadFile("fixtures/v2s1-invalid-signatures.manifest.json")
	require.NoError(t, err)
	sig, err = VerifyDockerManifestSignature(signature, invalidManifest, TestImageSignatureReference, mech, TestKeyFingerprint)
	assert.Error(t, err)
	assert.Nil(t, sig)

	// Error verifying signature
	corruptSignature, err := ioutil.ReadFile("fixtures/corrupt.signature")
	sig, err = VerifyDockerManifestSignature(corruptSignature, manifest, TestImageSignatureReference, mech, TestKeyFingerprint)
	assert.Error(t, err)
	assert.Nil(t, sig)

	// Key fingerprint mismatch
	sig, err = VerifyDockerManifestSignature(signature, manifest, TestImageSignatureReference, mech, "unexpected fingerprint")
	assert.Error(t, err)
	assert.Nil(t, sig)

	// Invalid reference in the signature
	invalidReferenceSignature, err := ioutil.ReadFile("fixtures/invalid-reference.signature")
	sig, err = VerifyDockerManifestSignature(invalidReferenceSignature, manifest, TestImageSignatureReference, mech, TestKeyFingerprint)
	assert.Error(t, err)
	assert.Nil(t, sig)

	// Docker reference mismatch
	sig, err = VerifyDockerManifestSignature(signature, manifest, "example.com/doesnt/match", mech, TestKeyFingerprint)
	assert.Error(t, err)
	assert.Nil(t, sig)

	// Docker manifest digest mismatch
	sig, err = VerifyDockerManifestSignature(signature, []byte("unexpected manifest"), TestImageSignatureReference, mech, TestKeyFingerprint)
	assert.Error(t, err)
	assert.Nil(t, sig)
}
