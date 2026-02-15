package keys

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"sync"
)

type Algorithm string

const (
	RS256 Algorithm = "RS256"
	ES256 Algorithm = "ES256"
)

type KeyConfig struct {
	ID             string
	Algorithm      Algorithm
	PrivateKeyPath string
	PublicKeyPath  string
}

type Manager struct {
	keys       map[string]*SigningKey
	currentKID string
	mu         sync.RWMutex
}

type SigningKey struct {
	KID        string
	Algorithm  Algorithm
	PrivateKey interface{}
	PublicKey  interface{}
}

type JWKS struct {
	Keys []JWK `json:"keys"`
}

type JWK struct {
	KTY string `json:"kty"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	KID string `json:"kid"`

	N string `json:"n,omitempty"`
	E string `json:"e,omitempty"`

	Crv string `json:"crv,omitempty"`
	X   string `json:"x,omitempty"`
	Y   string `json:"y,omitempty"`
}

func NewManager(configs []KeyConfig, currentKID string) (*Manager, error) {
	if len(configs) == 0 {
		return nil, fmt.Errorf("at least one key configuration is required")
	}

	manager := &Manager{
		keys:       make(map[string]*SigningKey),
		currentKID: currentKID,
	}

	for _, config := range configs {
		signingKey, err := loadSigningKey(config)
		if err != nil {
			return nil, fmt.Errorf("failed to load key %s: %w", config.ID, err)
		}
		manager.keys[config.ID] = signingKey
	}

	if _, exists := manager.keys[currentKID]; !exists {
		return nil, fmt.Errorf("current key ID %s not found in loaded keys", currentKID)
	}

	return manager, nil
}

func loadSigningKey(config KeyConfig) (*SigningKey, error) {
	var privateKey interface{}
	var publicKey interface{}
	var err error

	switch config.Algorithm {
	case RS256:
		privateKey, err = LoadRSAPrivateKey(config.PrivateKeyPath)
		if err != nil {
			return nil, err
		}
		rsaPrivateKey := privateKey.(*rsa.PrivateKey)
		publicKey = &rsaPrivateKey.PublicKey

	case ES256:
		privateKey, err = LoadECDSAPrivateKey(config.PrivateKeyPath)
		if err != nil {
			return nil, err
		}
		ecdsaPrivateKey := privateKey.(*ecdsa.PrivateKey)
		publicKey = &ecdsaPrivateKey.PublicKey

	default:
		return nil, fmt.Errorf("unsupported algorithm: %s", config.Algorithm)
	}

	return &SigningKey{
		KID:        config.ID,
		Algorithm:  config.Algorithm,
		PrivateKey: privateKey,
		PublicKey:  publicKey,
	}, nil
}

func (m *Manager) GetCurrentKey() (*SigningKey, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	key, exists := m.keys[m.currentKID]
	if !exists {
		return nil, fmt.Errorf("current key %s not found", m.currentKID)
	}

	return key, nil
}

func (m *Manager) GetKey(kid string) (*SigningKey, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	key, exists := m.keys[kid]
	if !exists {
		return nil, fmt.Errorf("key %s not found", kid)
	}

	return key, nil
}

func (m *Manager) GetJWKS() *JWKS {
	m.mu.RLock()
	defer m.mu.RUnlock()

	jwks := &JWKS{
		Keys: make([]JWK, 0, len(m.keys)),
	}

	for _, key := range m.keys {
		var jwk JWK

		switch key.Algorithm {
		case RS256:
			rsaPublicKey := key.PublicKey.(*rsa.PublicKey)
			jwk = RSAPublicKeyToJWK(rsaPublicKey, key.KID, string(key.Algorithm))
		case ES256:
			ecdsaPublicKey := key.PublicKey.(*ecdsa.PublicKey)
			jwk = ECDSAPublicKeyToJWK(ecdsaPublicKey, key.KID, string(key.Algorithm))
		}

		jwks.Keys = append(jwks.Keys, jwk)
	}

	return jwks
}

func (m *Manager) Sign(data []byte) ([]byte, string, error) {
	key, err := m.GetCurrentKey()
	if err != nil {
		return nil, "", err
	}

	signature, err := signData(data, key.PrivateKey, key.Algorithm)
	if err != nil {
		return nil, "", fmt.Errorf("failed to sign data: %w", err)
	}

	return signature, key.KID, nil
}

func (m *Manager) Verify(data, signature []byte, kid string) error {
	key, err := m.GetKey(kid)
	if err != nil {
		return err
	}

	return verifySignature(data, signature, key.PublicKey, key.Algorithm)
}

func signData(data []byte, privateKey interface{}, algorithm Algorithm) ([]byte, error) {
	hash := sha256.Sum256(data)

	switch algorithm {
	case RS256:
		rsaKey, ok := privateKey.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("invalid RSA private key")
		}
		return rsa.SignPKCS1v15(rand.Reader, rsaKey, crypto.SHA256, hash[:])

	case ES256:
		ecdsaKey, ok := privateKey.(*ecdsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("invalid ECDSA private key")
		}
		r, s, err := ecdsa.Sign(rand.Reader, ecdsaKey, hash[:])
		if err != nil {
			return nil, err
		}
		rBytes := r.Bytes()
		sBytes := s.Bytes()
		signature := make([]byte, 64)
		copy(signature[32-len(rBytes):32], rBytes)
		copy(signature[64-len(sBytes):64], sBytes)
		return signature, nil

	default:
		return nil, fmt.Errorf("unsupported algorithm: %s", algorithm)
	}
}

func verifySignature(data, signature []byte, publicKey interface{}, algorithm Algorithm) error {
	hash := sha256.Sum256(data)

	switch algorithm {
	case RS256:
		rsaKey, ok := publicKey.(*rsa.PublicKey)
		if !ok {
			return fmt.Errorf("invalid RSA public key")
		}
		return rsa.VerifyPKCS1v15(rsaKey, crypto.SHA256, hash[:], signature)

	case ES256:
		ecdsaKey, ok := publicKey.(*ecdsa.PublicKey)
		if !ok {
			return fmt.Errorf("invalid ECDSA public key")
		}
		if len(signature) != 64 {
			return fmt.Errorf("invalid ECDSA signature length")
		}
		r := new(big.Int).SetBytes(signature[:32])
		s := new(big.Int).SetBytes(signature[32:])
		if !ecdsa.Verify(ecdsaKey, hash[:], r, s) {
			return fmt.Errorf("signature verification failed")
		}
		return nil

	default:
		return fmt.Errorf("unsupported algorithm: %s", algorithm)
	}
}

func LoadRSAPrivateKey(path string) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read key file: %w", err)
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		parsedKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse RSA private key: %w", err)
		}
		rsaKey, ok := parsedKey.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("not an RSA private key")
		}
		return rsaKey, nil
	}

	return key, nil
}

func LoadECDSAPrivateKey(path string) (*ecdsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read key file: %w", err)
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	key, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		parsedKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse ECDSA private key: %w", err)
		}
		ecdsaKey, ok := parsedKey.(*ecdsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("not an ECDSA private key")
		}
		return ecdsaKey, nil
	}

	return key, nil
}

func RSAPublicKeyToJWK(key *rsa.PublicKey, kid, alg string) JWK {
	return JWK{
		KTY: "RSA",
		Use: "sig",
		Alg: alg,
		KID: kid,
		N:   base64URLEncode(key.N.Bytes()),
		E:   base64URLEncode(big.NewInt(int64(key.E)).Bytes()),
	}
}

func ECDSAPublicKeyToJWK(key *ecdsa.PublicKey, kid, alg string) JWK {
	return JWK{
		KTY: "EC",
		Use: "sig",
		Alg: alg,
		KID: kid,
		Crv: "P-256",
		X:   base64URLEncode(key.X.Bytes()),
		Y:   base64URLEncode(key.Y.Bytes()),
	}
}

func base64URLEncode(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}
