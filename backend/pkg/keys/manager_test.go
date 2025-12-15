package keys

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
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

func TestBase64URLEncode(t *testing.T) {
	data := []byte("hello world")
	encoded := base64URLEncode(data)

	assert.NotContains(t, encoded, "=")
	assert.NotEmpty(t, encoded)
}
