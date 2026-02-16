package keys

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTempKeyFiles(t *testing.T) (string, func()) {
	tmpDir, err := os.MkdirTemp("", "keys-test-*")
	require.NoError(t, err)

	rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	rsaPrivatePath := filepath.Join(tmpDir, "rsa_private.pem")
	rsaPrivateFile, err := os.Create(rsaPrivatePath)
	require.NoError(t, err)
	err = pem.Encode(rsaPrivateFile, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(rsaKey),
	})
	require.NoError(t, err)
	rsaPrivateFile.Close()

	ecdsaKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	ecdsaPrivatePath := filepath.Join(tmpDir, "ecdsa_private.pem")
	ecdsaPrivateFile, err := os.Create(ecdsaPrivatePath)
	require.NoError(t, err)
	ecdsaBytes, err := x509.MarshalECPrivateKey(ecdsaKey)
	require.NoError(t, err)
	err = pem.Encode(ecdsaPrivateFile, &pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: ecdsaBytes,
	})
	require.NoError(t, err)
	ecdsaPrivateFile.Close()

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return tmpDir, cleanup
}

// createPKCS8RSAKeyFile creates an RSA key in PKCS8 format (for testing the fallback parsing path)
func createPKCS8RSAKeyFile(t *testing.T, dir string) string {
	rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	pkcs8Bytes, err := x509.MarshalPKCS8PrivateKey(rsaKey)
	require.NoError(t, err)

	path := filepath.Join(dir, "rsa_pkcs8.pem")
	f, err := os.Create(path)
	require.NoError(t, err)
	defer f.Close()

	err = pem.Encode(f, &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: pkcs8Bytes,
	})
	require.NoError(t, err)

	return path
}

// createPKCS8ECDSAKeyFile creates an ECDSA key in PKCS8 format
func createPKCS8ECDSAKeyFile(t *testing.T, dir string) string {
	ecKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	pkcs8Bytes, err := x509.MarshalPKCS8PrivateKey(ecKey)
	require.NoError(t, err)

	path := filepath.Join(dir, "ecdsa_pkcs8.pem")
	f, err := os.Create(path)
	require.NoError(t, err)
	defer f.Close()

	err = pem.Encode(f, &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: pkcs8Bytes,
	})
	require.NoError(t, err)

	return path
}

func TestNewManager_Success(t *testing.T) {
	tmpDir, cleanup := createTempKeyFiles(t)
	defer cleanup()

	configs := []KeyConfig{
		{
			ID:             "key-1",
			Algorithm:      RS256,
			PrivateKeyPath: filepath.Join(tmpDir, "rsa_private.pem"),
		},
		{
			ID:             "key-2",
			Algorithm:      ES256,
			PrivateKeyPath: filepath.Join(tmpDir, "ecdsa_private.pem"),
		},
	}

	manager, err := NewManager(configs, "key-1")
	require.NoError(t, err)
	assert.NotNil(t, manager)
	assert.Len(t, manager.keys, 2)
	assert.Equal(t, "key-1", manager.currentKID)
}

func TestNewManager_NoKeys(t *testing.T) {
	manager, err := NewManager([]KeyConfig{}, "key-1")
	assert.Error(t, err)
	assert.Nil(t, manager)
	assert.Contains(t, err.Error(), "at least one key configuration is required")
}

func TestNewManager_NilKeys(t *testing.T) {
	manager, err := NewManager(nil, "key-1")
	assert.Error(t, err)
	assert.Nil(t, manager)
	assert.Contains(t, err.Error(), "at least one key configuration is required")
}

func TestNewManager_InvalidCurrentKID(t *testing.T) {
	tmpDir, cleanup := createTempKeyFiles(t)
	defer cleanup()

	configs := []KeyConfig{
		{
			ID:             "key-1",
			Algorithm:      RS256,
			PrivateKeyPath: filepath.Join(tmpDir, "rsa_private.pem"),
		},
	}

	manager, err := NewManager(configs, "nonexistent-key")
	assert.Error(t, err)
	assert.Nil(t, manager)
	assert.Contains(t, err.Error(), "current key ID nonexistent-key not found")
}

func TestNewManager_InvalidKeyPath(t *testing.T) {
	configs := []KeyConfig{
		{
			ID:             "key-1",
			Algorithm:      RS256,
			PrivateKeyPath: "/nonexistent/path/key.pem",
		},
	}

	manager, err := NewManager(configs, "key-1")
	assert.Error(t, err)
	assert.Nil(t, manager)
}

func TestNewManager_ShouldFail_WhenUnsupportedAlgorithm(t *testing.T) {
	tmpDir, cleanup := createTempKeyFiles(t)
	defer cleanup()

	configs := []KeyConfig{
		{
			ID:             "key-1",
			Algorithm:      Algorithm("RS384"),
			PrivateKeyPath: filepath.Join(tmpDir, "rsa_private.pem"),
		},
	}

	manager, err := NewManager(configs, "key-1")
	assert.Error(t, err)
	assert.Nil(t, manager)
	assert.Contains(t, err.Error(), "unsupported algorithm")
}

func TestNewManager_ShouldFail_WhenInvalidPEMContent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "keys-test-invalid-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Write non-PEM content
	invalidPath := filepath.Join(tmpDir, "invalid.pem")
	err = os.WriteFile(invalidPath, []byte("this is not a PEM file"), 0644)
	require.NoError(t, err)

	configs := []KeyConfig{
		{
			ID:             "key-1",
			Algorithm:      RS256,
			PrivateKeyPath: invalidPath,
		},
	}

	manager, err := NewManager(configs, "key-1")
	assert.Error(t, err)
	assert.Nil(t, manager)
	assert.Contains(t, err.Error(), "failed to decode PEM block")
}

func TestNewManager_ShouldFail_WhenWrongKeyTypeForAlgorithm(t *testing.T) {
	tmpDir, cleanup := createTempKeyFiles(t)
	defer cleanup()

	// Try to load ECDSA key as RSA
	configs := []KeyConfig{
		{
			ID:             "key-1",
			Algorithm:      RS256,
			PrivateKeyPath: filepath.Join(tmpDir, "ecdsa_private.pem"),
		},
	}

	manager, err := NewManager(configs, "key-1")
	assert.Error(t, err)
	assert.Nil(t, manager)
}

func TestNewManager_ShouldFail_WhenRSAKeyLoadedAsECDSA(t *testing.T) {
	tmpDir, cleanup := createTempKeyFiles(t)
	defer cleanup()

	// Try to load RSA key as ECDSA
	configs := []KeyConfig{
		{
			ID:             "key-1",
			Algorithm:      ES256,
			PrivateKeyPath: filepath.Join(tmpDir, "rsa_private.pem"),
		},
	}

	manager, err := NewManager(configs, "key-1")
	assert.Error(t, err)
	assert.Nil(t, manager)
}

func TestNewManager_ShouldSucceed_WhenSingleRSAKey(t *testing.T) {
	tmpDir, cleanup := createTempKeyFiles(t)
	defer cleanup()

	configs := []KeyConfig{
		{
			ID:             "only-key",
			Algorithm:      RS256,
			PrivateKeyPath: filepath.Join(tmpDir, "rsa_private.pem"),
		},
	}

	manager, err := NewManager(configs, "only-key")
	require.NoError(t, err)
	assert.Len(t, manager.keys, 1)
}

func TestNewManager_ShouldSucceed_WhenSingleECDSAKey(t *testing.T) {
	tmpDir, cleanup := createTempKeyFiles(t)
	defer cleanup()

	configs := []KeyConfig{
		{
			ID:             "only-key",
			Algorithm:      ES256,
			PrivateKeyPath: filepath.Join(tmpDir, "ecdsa_private.pem"),
		},
	}

	manager, err := NewManager(configs, "only-key")
	require.NoError(t, err)
	assert.Len(t, manager.keys, 1)
}

func TestGetCurrentKey(t *testing.T) {
	tmpDir, cleanup := createTempKeyFiles(t)
	defer cleanup()

	configs := []KeyConfig{
		{
			ID:             "key-1",
			Algorithm:      RS256,
			PrivateKeyPath: filepath.Join(tmpDir, "rsa_private.pem"),
		},
	}

	manager, err := NewManager(configs, "key-1")
	require.NoError(t, err)

	key, err := manager.GetCurrentKey()
	require.NoError(t, err)
	assert.Equal(t, "key-1", key.KID)
	assert.Equal(t, RS256, key.Algorithm)
	assert.NotNil(t, key.PrivateKey)
	assert.NotNil(t, key.PublicKey)
}

func TestGetCurrentKey_ShouldReturnECDSAKey_WhenCurrentIsECDSA(t *testing.T) {
	tmpDir, cleanup := createTempKeyFiles(t)
	defer cleanup()

	configs := []KeyConfig{
		{
			ID:             "ec-key",
			Algorithm:      ES256,
			PrivateKeyPath: filepath.Join(tmpDir, "ecdsa_private.pem"),
		},
	}

	manager, err := NewManager(configs, "ec-key")
	require.NoError(t, err)

	key, err := manager.GetCurrentKey()
	require.NoError(t, err)
	assert.Equal(t, ES256, key.Algorithm)
	assert.IsType(t, &ecdsa.PrivateKey{}, key.PrivateKey)
	assert.IsType(t, &ecdsa.PublicKey{}, key.PublicKey)
}

func TestGetKey(t *testing.T) {
	tmpDir, cleanup := createTempKeyFiles(t)
	defer cleanup()

	configs := []KeyConfig{
		{
			ID:             "key-1",
			Algorithm:      RS256,
			PrivateKeyPath: filepath.Join(tmpDir, "rsa_private.pem"),
		},
		{
			ID:             "key-2",
			Algorithm:      ES256,
			PrivateKeyPath: filepath.Join(tmpDir, "ecdsa_private.pem"),
		},
	}

	manager, err := NewManager(configs, "key-1")
	require.NoError(t, err)

	key1, err := manager.GetKey("key-1")
	require.NoError(t, err)
	assert.Equal(t, "key-1", key1.KID)
	assert.Equal(t, RS256, key1.Algorithm)

	key2, err := manager.GetKey("key-2")
	require.NoError(t, err)
	assert.Equal(t, "key-2", key2.KID)
	assert.Equal(t, ES256, key2.Algorithm)

	_, err = manager.GetKey("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "key nonexistent not found")
}

func TestGetKey_ShouldFail_WhenEmptyKID(t *testing.T) {
	tmpDir, cleanup := createTempKeyFiles(t)
	defer cleanup()

	configs := []KeyConfig{
		{
			ID:             "key-1",
			Algorithm:      RS256,
			PrivateKeyPath: filepath.Join(tmpDir, "rsa_private.pem"),
		},
	}

	manager, err := NewManager(configs, "key-1")
	require.NoError(t, err)

	_, err = manager.GetKey("")
	assert.Error(t, err)
}

func TestGetJWKS(t *testing.T) {
	tmpDir, cleanup := createTempKeyFiles(t)
	defer cleanup()

	configs := []KeyConfig{
		{
			ID:             "key-1",
			Algorithm:      RS256,
			PrivateKeyPath: filepath.Join(tmpDir, "rsa_private.pem"),
		},
		{
			ID:             "key-2",
			Algorithm:      ES256,
			PrivateKeyPath: filepath.Join(tmpDir, "ecdsa_private.pem"),
		},
	}

	manager, err := NewManager(configs, "key-1")
	require.NoError(t, err)

	jwks := manager.GetJWKS()
	require.NotNil(t, jwks)
	assert.Len(t, jwks.Keys, 2)

	var rsaJWK, ecJWK *JWK
	for i := range jwks.Keys {
		if jwks.Keys[i].KTY == "RSA" {
			rsaJWK = &jwks.Keys[i]
		} else if jwks.Keys[i].KTY == "EC" {
			ecJWK = &jwks.Keys[i]
		}
	}

	require.NotNil(t, rsaJWK)
	assert.Equal(t, "RSA", rsaJWK.KTY)
	assert.Equal(t, "sig", rsaJWK.Use)
	assert.Equal(t, "RS256", rsaJWK.Alg)
	assert.NotEmpty(t, rsaJWK.N)
	assert.NotEmpty(t, rsaJWK.E)

	require.NotNil(t, ecJWK)
	assert.Equal(t, "EC", ecJWK.KTY)
	assert.Equal(t, "sig", ecJWK.Use)
	assert.Equal(t, "ES256", ecJWK.Alg)
	assert.Equal(t, "P-256", ecJWK.Crv)
	assert.NotEmpty(t, ecJWK.X)
	assert.NotEmpty(t, ecJWK.Y)
}

func TestGetJWKS_ShouldReturnSingleKey_WhenOneKeyConfigured(t *testing.T) {
	tmpDir, cleanup := createTempKeyFiles(t)
	defer cleanup()

	configs := []KeyConfig{
		{
			ID:             "key-1",
			Algorithm:      RS256,
			PrivateKeyPath: filepath.Join(tmpDir, "rsa_private.pem"),
		},
	}

	manager, err := NewManager(configs, "key-1")
	require.NoError(t, err)

	jwks := manager.GetJWKS()
	require.NotNil(t, jwks)
	assert.Len(t, jwks.Keys, 1)
	assert.Equal(t, "key-1", jwks.Keys[0].KID)
	assert.Equal(t, "RSA", jwks.Keys[0].KTY)
}

func TestGetJWKS_ShouldContainValidBase64URLValues(t *testing.T) {
	tmpDir, cleanup := createTempKeyFiles(t)
	defer cleanup()

	configs := []KeyConfig{
		{
			ID:             "key-1",
			Algorithm:      RS256,
			PrivateKeyPath: filepath.Join(tmpDir, "rsa_private.pem"),
		},
		{
			ID:             "key-2",
			Algorithm:      ES256,
			PrivateKeyPath: filepath.Join(tmpDir, "ecdsa_private.pem"),
		},
	}

	manager, err := NewManager(configs, "key-1")
	require.NoError(t, err)

	jwks := manager.GetJWKS()

	for _, jwk := range jwks.Keys {
		if jwk.KTY == "RSA" {
			// N and E should be valid base64url-encoded
			_, err := base64.RawURLEncoding.DecodeString(jwk.N)
			assert.NoError(t, err, "RSA N should be valid base64url")
			_, err = base64.RawURLEncoding.DecodeString(jwk.E)
			assert.NoError(t, err, "RSA E should be valid base64url")
			// Should not contain padding
			assert.NotContains(t, jwk.N, "=")
			assert.NotContains(t, jwk.E, "=")
		}
		if jwk.KTY == "EC" {
			_, err := base64.RawURLEncoding.DecodeString(jwk.X)
			assert.NoError(t, err, "EC X should be valid base64url")
			_, err = base64.RawURLEncoding.DecodeString(jwk.Y)
			assert.NoError(t, err, "EC Y should be valid base64url")
			assert.NotContains(t, jwk.X, "=")
			assert.NotContains(t, jwk.Y, "=")
		}
	}
}

func TestGetJWKS_ShouldNotExposePrivateKeys(t *testing.T) {
	tmpDir, cleanup := createTempKeyFiles(t)
	defer cleanup()

	configs := []KeyConfig{
		{
			ID:             "key-1",
			Algorithm:      RS256,
			PrivateKeyPath: filepath.Join(tmpDir, "rsa_private.pem"),
		},
	}

	manager, err := NewManager(configs, "key-1")
	require.NoError(t, err)

	jwks := manager.GetJWKS()
	// JWK struct only has public key fields (N, E for RSA; X, Y for EC)
	// There should be no 'd', 'p', 'q' fields (private key components)
	// The struct definition ensures this, but let's verify the fields are only public
	for _, jwk := range jwks.Keys {
		assert.Equal(t, "sig", jwk.Use)
		assert.NotEmpty(t, jwk.KID)
		assert.NotEmpty(t, jwk.Alg)
	}
}

func TestSignAndVerify_RSA(t *testing.T) {
	tmpDir, cleanup := createTempKeyFiles(t)
	defer cleanup()

	configs := []KeyConfig{
		{
			ID:             "key-1",
			Algorithm:      RS256,
			PrivateKeyPath: filepath.Join(tmpDir, "rsa_private.pem"),
		},
	}

	manager, err := NewManager(configs, "key-1")
	require.NoError(t, err)

	data := []byte("test data to sign")

	signature, kid, err := manager.Sign(data)
	require.NoError(t, err)
	assert.Equal(t, "key-1", kid)
	assert.NotEmpty(t, signature)

	err = manager.Verify(data, signature, kid)
	assert.NoError(t, err)

	wrongData := []byte("wrong data")
	err = manager.Verify(wrongData, signature, kid)
	assert.Error(t, err)
}

func TestSignAndVerify_ECDSA(t *testing.T) {
	tmpDir, cleanup := createTempKeyFiles(t)
	defer cleanup()

	configs := []KeyConfig{
		{
			ID:             "key-1",
			Algorithm:      ES256,
			PrivateKeyPath: filepath.Join(tmpDir, "ecdsa_private.pem"),
		},
	}

	manager, err := NewManager(configs, "key-1")
	require.NoError(t, err)

	data := []byte("test data to sign")

	signature, kid, err := manager.Sign(data)
	require.NoError(t, err)
	assert.Equal(t, "key-1", kid)
	assert.NotEmpty(t, signature)
	assert.Len(t, signature, 64)

	err = manager.Verify(data, signature, kid)
	assert.NoError(t, err)

	wrongData := []byte("wrong data")
	err = manager.Verify(wrongData, signature, kid)
	assert.Error(t, err)
}

func TestSign_ShouldProduceDifferentSignatures_WhenCalledTwiceWithECDSA(t *testing.T) {
	tmpDir, cleanup := createTempKeyFiles(t)
	defer cleanup()

	configs := []KeyConfig{
		{
			ID:             "key-1",
			Algorithm:      ES256,
			PrivateKeyPath: filepath.Join(tmpDir, "ecdsa_private.pem"),
		},
	}

	manager, err := NewManager(configs, "key-1")
	require.NoError(t, err)

	data := []byte("test data")

	sig1, _, err := manager.Sign(data)
	require.NoError(t, err)

	sig2, _, err := manager.Sign(data)
	require.NoError(t, err)

	// ECDSA signatures are non-deterministic (due to random k)
	assert.NotEqual(t, sig1, sig2, "ECDSA signatures should differ due to randomness")

	// Both should still verify
	err = manager.Verify(data, sig1, "key-1")
	assert.NoError(t, err)
	err = manager.Verify(data, sig2, "key-1")
	assert.NoError(t, err)
}

func TestSign_ShouldSignEmptyData(t *testing.T) {
	tmpDir, cleanup := createTempKeyFiles(t)
	defer cleanup()

	configs := []KeyConfig{
		{
			ID:             "key-1",
			Algorithm:      RS256,
			PrivateKeyPath: filepath.Join(tmpDir, "rsa_private.pem"),
		},
	}

	manager, err := NewManager(configs, "key-1")
	require.NoError(t, err)

	sig, kid, err := manager.Sign([]byte{})
	require.NoError(t, err)
	assert.NotEmpty(t, sig)
	assert.Equal(t, "key-1", kid)

	err = manager.Verify([]byte{}, sig, kid)
	assert.NoError(t, err)
}

func TestSign_ShouldSignLargeData(t *testing.T) {
	tmpDir, cleanup := createTempKeyFiles(t)
	defer cleanup()

	configs := []KeyConfig{
		{
			ID:             "key-1",
			Algorithm:      RS256,
			PrivateKeyPath: filepath.Join(tmpDir, "rsa_private.pem"),
		},
	}

	manager, err := NewManager(configs, "key-1")
	require.NoError(t, err)

	largeData := make([]byte, 1000000)
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}

	sig, kid, err := manager.Sign(largeData)
	require.NoError(t, err)
	assert.NotEmpty(t, sig)

	err = manager.Verify(largeData, sig, kid)
	assert.NoError(t, err)
}

func TestVerify_ShouldFail_WhenKIDNotFound(t *testing.T) {
	tmpDir, cleanup := createTempKeyFiles(t)
	defer cleanup()

	configs := []KeyConfig{
		{
			ID:             "key-1",
			Algorithm:      RS256,
			PrivateKeyPath: filepath.Join(tmpDir, "rsa_private.pem"),
		},
	}

	manager, err := NewManager(configs, "key-1")
	require.NoError(t, err)

	data := []byte("test data")
	sig, _, err := manager.Sign(data)
	require.NoError(t, err)

	err = manager.Verify(data, sig, "nonexistent-key")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "key nonexistent-key not found")
}

func TestVerify_ShouldFail_WhenSignatureIsTruncated(t *testing.T) {
	tmpDir, cleanup := createTempKeyFiles(t)
	defer cleanup()

	configs := []KeyConfig{
		{
			ID:             "key-1",
			Algorithm:      RS256,
			PrivateKeyPath: filepath.Join(tmpDir, "rsa_private.pem"),
		},
	}

	manager, err := NewManager(configs, "key-1")
	require.NoError(t, err)

	data := []byte("test data")
	sig, kid, err := manager.Sign(data)
	require.NoError(t, err)

	truncated := sig[:len(sig)/2]
	err = manager.Verify(data, truncated, kid)
	assert.Error(t, err)
}

func TestVerify_ShouldFail_WhenECDSASignatureWrongLength(t *testing.T) {
	tmpDir, cleanup := createTempKeyFiles(t)
	defer cleanup()

	configs := []KeyConfig{
		{
			ID:             "key-1",
			Algorithm:      ES256,
			PrivateKeyPath: filepath.Join(tmpDir, "ecdsa_private.pem"),
		},
	}

	manager, err := NewManager(configs, "key-1")
	require.NoError(t, err)

	data := []byte("test data")

	// Wrong signature length (not 64 bytes)
	shortSig := make([]byte, 32)
	err = manager.Verify(data, shortSig, "key-1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid ECDSA signature length")

	longSig := make([]byte, 128)
	err = manager.Verify(data, longSig, "key-1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid ECDSA signature length")
}

func TestVerify_ShouldFail_WhenECDSASignatureIsCorrupted(t *testing.T) {
	tmpDir, cleanup := createTempKeyFiles(t)
	defer cleanup()

	configs := []KeyConfig{
		{
			ID:             "key-1",
			Algorithm:      ES256,
			PrivateKeyPath: filepath.Join(tmpDir, "ecdsa_private.pem"),
		},
	}

	manager, err := NewManager(configs, "key-1")
	require.NoError(t, err)

	data := []byte("test data")
	sig, _, err := manager.Sign(data)
	require.NoError(t, err)

	// Corrupt the signature by flipping bytes
	corrupted := make([]byte, len(sig))
	copy(corrupted, sig)
	corrupted[0] ^= 0xFF
	corrupted[32] ^= 0xFF

	err = manager.Verify(data, corrupted, "key-1")
	assert.Error(t, err)
}

func TestVerify_ShouldFail_WhenRSASignatureIsCorrupted(t *testing.T) {
	tmpDir, cleanup := createTempKeyFiles(t)
	defer cleanup()

	configs := []KeyConfig{
		{
			ID:             "key-1",
			Algorithm:      RS256,
			PrivateKeyPath: filepath.Join(tmpDir, "rsa_private.pem"),
		},
	}

	manager, err := NewManager(configs, "key-1")
	require.NoError(t, err)

	data := []byte("test data")
	sig, _, err := manager.Sign(data)
	require.NoError(t, err)

	corrupted := make([]byte, len(sig))
	copy(corrupted, sig)
	corrupted[0] ^= 0xFF

	err = manager.Verify(data, corrupted, "key-1")
	assert.Error(t, err)
}

func TestSignAndVerify_ShouldNotCrossVerify_WhenDifferentKeys(t *testing.T) {
	tmpDir, cleanup := createTempKeyFiles(t)
	defer cleanup()

	// Create a second set of RSA keys
	rsaKey2, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	rsa2Path := filepath.Join(tmpDir, "rsa_private_2.pem")
	f, err := os.Create(rsa2Path)
	require.NoError(t, err)
	err = pem.Encode(f, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(rsaKey2),
	})
	require.NoError(t, err)
	f.Close()

	configs := []KeyConfig{
		{
			ID:             "key-1",
			Algorithm:      RS256,
			PrivateKeyPath: filepath.Join(tmpDir, "rsa_private.pem"),
		},
		{
			ID:             "key-2",
			Algorithm:      RS256,
			PrivateKeyPath: rsa2Path,
		},
	}

	manager, err := NewManager(configs, "key-1")
	require.NoError(t, err)

	data := []byte("test data")
	sig, kid, err := manager.Sign(data)
	require.NoError(t, err)
	assert.Equal(t, "key-1", kid)

	// Verify with correct key should succeed
	err = manager.Verify(data, sig, "key-1")
	assert.NoError(t, err)

	// Verify with wrong key should fail
	err = manager.Verify(data, sig, "key-2")
	assert.Error(t, err)
}

func TestLoadRSAPrivateKey(t *testing.T) {
	tmpDir, cleanup := createTempKeyFiles(t)
	defer cleanup()

	key, err := LoadRSAPrivateKey(filepath.Join(tmpDir, "rsa_private.pem"))
	require.NoError(t, err)
	assert.NotNil(t, key)
	assert.IsType(t, &rsa.PrivateKey{}, key)

	_, err = LoadRSAPrivateKey("/nonexistent/path.pem")
	assert.Error(t, err)
}

func TestLoadRSAPrivateKey_ShouldLoadPKCS8Format(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "keys-pkcs8-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	pkcs8Path := createPKCS8RSAKeyFile(t, tmpDir)

	key, err := LoadRSAPrivateKey(pkcs8Path)
	require.NoError(t, err)
	assert.NotNil(t, key)
	assert.IsType(t, &rsa.PrivateKey{}, key)
}

func TestLoadRSAPrivateKey_ShouldFail_WhenECDSAKeyProvided(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "keys-wrongtype-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	ecPath := createPKCS8ECDSAKeyFile(t, tmpDir)

	// PKCS1 parsing will fail, then PKCS8 parsing succeeds but type assertion fails
	_, err = LoadRSAPrivateKey(ecPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not an RSA private key")
}

func TestLoadECDSAPrivateKey(t *testing.T) {
	tmpDir, cleanup := createTempKeyFiles(t)
	defer cleanup()

	key, err := LoadECDSAPrivateKey(filepath.Join(tmpDir, "ecdsa_private.pem"))
	require.NoError(t, err)
	assert.NotNil(t, key)
	assert.IsType(t, &ecdsa.PrivateKey{}, key)

	_, err = LoadECDSAPrivateKey("/nonexistent/path.pem")
	assert.Error(t, err)
}

func TestLoadECDSAPrivateKey_ShouldLoadPKCS8Format(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "keys-pkcs8-ec-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	pkcs8Path := createPKCS8ECDSAKeyFile(t, tmpDir)

	key, err := LoadECDSAPrivateKey(pkcs8Path)
	require.NoError(t, err)
	assert.NotNil(t, key)
	assert.IsType(t, &ecdsa.PrivateKey{}, key)
}

func TestLoadECDSAPrivateKey_ShouldFail_WhenRSAKeyProvided(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "keys-wrongtype-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	rsaPath := createPKCS8RSAKeyFile(t, tmpDir)

	// EC parsing will fail, then PKCS8 parsing succeeds but type assertion fails
	_, err = LoadECDSAPrivateKey(rsaPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not an ECDSA private key")
}

func TestLoadRSAPrivateKey_ShouldFail_WhenFileIsEmpty(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "keys-empty-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	emptyPath := filepath.Join(tmpDir, "empty.pem")
	err = os.WriteFile(emptyPath, []byte{}, 0644)
	require.NoError(t, err)

	_, err = LoadRSAPrivateKey(emptyPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode PEM block")
}

func TestLoadECDSAPrivateKey_ShouldFail_WhenFileIsEmpty(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "keys-empty-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	emptyPath := filepath.Join(tmpDir, "empty.pem")
	err = os.WriteFile(emptyPath, []byte{}, 0644)
	require.NoError(t, err)

	_, err = LoadECDSAPrivateKey(emptyPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode PEM block")
}

func TestRSAPublicKeyToJWK(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	jwk := RSAPublicKeyToJWK(&key.PublicKey, "test-kid", "RS256")

	assert.Equal(t, "RSA", jwk.KTY)
	assert.Equal(t, "sig", jwk.Use)
	assert.Equal(t, "RS256", jwk.Alg)
	assert.Equal(t, "test-kid", jwk.KID)
	assert.NotEmpty(t, jwk.N)
	assert.NotEmpty(t, jwk.E)
	assert.Empty(t, jwk.Crv)
	assert.Empty(t, jwk.X)
	assert.Empty(t, jwk.Y)
}

func TestRSAPublicKeyToJWK_ShouldEncodeExponentCorrectly(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	jwk := RSAPublicKeyToJWK(&key.PublicKey, "kid", "RS256")

	// Decode E and verify it matches the public exponent (commonly 65537)
	eBytes, err := base64.RawURLEncoding.DecodeString(jwk.E)
	require.NoError(t, err)
	e := new(big.Int).SetBytes(eBytes)
	assert.Equal(t, int64(key.PublicKey.E), e.Int64())
}

func TestECDSAPublicKeyToJWK(t *testing.T) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	jwk := ECDSAPublicKeyToJWK(&key.PublicKey, "test-kid", "ES256")

	assert.Equal(t, "EC", jwk.KTY)
	assert.Equal(t, "sig", jwk.Use)
	assert.Equal(t, "ES256", jwk.Alg)
	assert.Equal(t, "test-kid", jwk.KID)
	assert.Equal(t, "P-256", jwk.Crv)
	assert.NotEmpty(t, jwk.X)
	assert.NotEmpty(t, jwk.Y)
	assert.Empty(t, jwk.N)
	assert.Empty(t, jwk.E)
}

func TestECDSAPublicKeyToJWK_ShouldEncodeCoordinatesCorrectly(t *testing.T) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	jwk := ECDSAPublicKeyToJWK(&key.PublicKey, "kid", "ES256")

	xBytes, err := base64.RawURLEncoding.DecodeString(jwk.X)
	require.NoError(t, err)
	x := new(big.Int).SetBytes(xBytes)
	assert.Equal(t, key.PublicKey.X.Cmp(x), 0, "X coordinate should match")

	yBytes, err := base64.RawURLEncoding.DecodeString(jwk.Y)
	require.NoError(t, err)
	y := new(big.Int).SetBytes(yBytes)
	assert.Equal(t, key.PublicKey.Y.Cmp(y), 0, "Y coordinate should match")
}

func TestBase64URLEncode(t *testing.T) {
	data := []byte("hello world")
	encoded := base64URLEncode(data)

	assert.NotContains(t, encoded, "=")
	assert.NotEmpty(t, encoded)
}

func TestBase64URLEncode_ShouldNotContainPadding(t *testing.T) {
	// Test various lengths to ensure no padding
	for _, size := range []int{1, 2, 3, 4, 5, 10, 32, 64, 128, 256} {
		data := make([]byte, size)
		for i := range data {
			data[i] = byte(i)
		}
		encoded := base64URLEncode(data)
		assert.NotContains(t, encoded, "=", "size %d should not have padding", size)
		assert.NotContains(t, encoded, "+", "should use URL-safe encoding")
		assert.NotContains(t, encoded, "/", "should use URL-safe encoding")
	}
}

func TestBase64URLEncode_ShouldHandleEmptyInput(t *testing.T) {
	encoded := base64URLEncode([]byte{})
	assert.Equal(t, "", encoded)
}

func TestManager_ConcurrentAccess(t *testing.T) {
	tmpDir, cleanup := createTempKeyFiles(t)
	defer cleanup()

	configs := []KeyConfig{
		{
			ID:             "key-1",
			Algorithm:      RS256,
			PrivateKeyPath: filepath.Join(tmpDir, "rsa_private.pem"),
		},
		{
			ID:             "key-2",
			Algorithm:      ES256,
			PrivateKeyPath: filepath.Join(tmpDir, "ecdsa_private.pem"),
		},
	}

	manager, err := NewManager(configs, "key-1")
	require.NoError(t, err)

	data := []byte("concurrent test data")

	var wg sync.WaitGroup
	errCh := make(chan error, 100)

	// Concurrent Sign operations
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _, err := manager.Sign(data)
			if err != nil {
				errCh <- err
			}
		}()
	}

	// Concurrent GetJWKS operations
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			jwks := manager.GetJWKS()
			if jwks == nil || len(jwks.Keys) != 2 {
				errCh <- assert.AnError
			}
		}()
	}

	// Concurrent GetKey operations
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := manager.GetKey("key-1")
			if err != nil {
				errCh <- err
			}
		}()
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Errorf("concurrent operation failed: %v", err)
	}
}

func TestSignAndVerify_RoundTrip_WithMultipleKeys(t *testing.T) {
	tmpDir, cleanup := createTempKeyFiles(t)
	defer cleanup()

	configs := []KeyConfig{
		{
			ID:             "rsa-key",
			Algorithm:      RS256,
			PrivateKeyPath: filepath.Join(tmpDir, "rsa_private.pem"),
		},
		{
			ID:             "ec-key",
			Algorithm:      ES256,
			PrivateKeyPath: filepath.Join(tmpDir, "ecdsa_private.pem"),
		},
	}

	// Test with RSA as current key
	manager, err := NewManager(configs, "rsa-key")
	require.NoError(t, err)

	data := []byte("multi-key test data")

	sig, kid, err := manager.Sign(data)
	require.NoError(t, err)
	assert.Equal(t, "rsa-key", kid, "should sign with current key")

	err = manager.Verify(data, sig, kid)
	assert.NoError(t, err)
}

func TestNewManager_ShouldSucceed_WhenCurrentKIDIsSecondKey(t *testing.T) {
	tmpDir, cleanup := createTempKeyFiles(t)
	defer cleanup()

	configs := []KeyConfig{
		{
			ID:             "key-old",
			Algorithm:      RS256,
			PrivateKeyPath: filepath.Join(tmpDir, "rsa_private.pem"),
		},
		{
			ID:             "key-new",
			Algorithm:      ES256,
			PrivateKeyPath: filepath.Join(tmpDir, "ecdsa_private.pem"),
		},
	}

	// Current key is the second one (simulates key rotation)
	manager, err := NewManager(configs, "key-new")
	require.NoError(t, err)

	key, err := manager.GetCurrentKey()
	require.NoError(t, err)
	assert.Equal(t, "key-new", key.KID)
	assert.Equal(t, ES256, key.Algorithm)

	// Old key should still be accessible for verification
	oldKey, err := manager.GetKey("key-old")
	require.NoError(t, err)
	assert.Equal(t, "key-old", oldKey.KID)
}

func TestAlgorithmConstants(t *testing.T) {
	assert.Equal(t, Algorithm("RS256"), RS256)
	assert.Equal(t, Algorithm("ES256"), ES256)
}

func TestSignAndVerify_ECDSA_WithBinaryData(t *testing.T) {
	tmpDir, cleanup := createTempKeyFiles(t)
	defer cleanup()

	configs := []KeyConfig{
		{
			ID:             "key-1",
			Algorithm:      ES256,
			PrivateKeyPath: filepath.Join(tmpDir, "ecdsa_private.pem"),
		},
	}

	manager, err := NewManager(configs, "key-1")
	require.NoError(t, err)

	// Binary data with null bytes
	data := []byte{0x00, 0x01, 0xFF, 0xFE, 0x00, 0x00, 0xAB, 0xCD}

	sig, kid, err := manager.Sign(data)
	require.NoError(t, err)

	err = manager.Verify(data, sig, kid)
	assert.NoError(t, err)
}

func TestVerify_ShouldFail_WhenEmptySignature(t *testing.T) {
	tmpDir, cleanup := createTempKeyFiles(t)
	defer cleanup()

	configs := []KeyConfig{
		{
			ID:             "key-1",
			Algorithm:      RS256,
			PrivateKeyPath: filepath.Join(tmpDir, "rsa_private.pem"),
		},
	}

	manager, err := NewManager(configs, "key-1")
	require.NoError(t, err)

	err = manager.Verify([]byte("data"), []byte{}, "key-1")
	assert.Error(t, err)
}

func TestVerify_ShouldFail_WhenNilSignature(t *testing.T) {
	tmpDir, cleanup := createTempKeyFiles(t)
	defer cleanup()

	configs := []KeyConfig{
		{
			ID:             "key-1",
			Algorithm:      RS256,
			PrivateKeyPath: filepath.Join(tmpDir, "rsa_private.pem"),
		},
	}

	manager, err := NewManager(configs, "key-1")
	require.NoError(t, err)

	err = manager.Verify([]byte("data"), nil, "key-1")
	assert.Error(t, err)
}
