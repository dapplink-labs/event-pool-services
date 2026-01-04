package common

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
)

type SIWEMessage struct {
	Domain    string
	Address   string
	URI       string
	Version   string
	ChainID   string
	Nonce     string
	IssuedAt  string
	Statement string
}

type JWTClaims struct {
	Address string `json:"address"`
	jwt.RegisteredClaims
}

type SIWEVerifier struct {
	jwtSecret []byte
	domain    string
}

func NewSIWEVerifier(jwtSecret, domain string) *SIWEVerifier {
	return &SIWEVerifier{
		jwtSecret: []byte(jwtSecret),
		domain:    domain,
	}
}

func (v *SIWEVerifier) ParseSIWEMessage(message string) (*SIWEMessage, error) {
	message = strings.TrimSpace(message)
	noncePart := strings.TrimPrefix(message, "Login to DappLink with nonce:")
	nonce := strings.TrimSpace(noncePart)

	if nonce == "" {
		return nil, errors.New("nonce is required in simplified format")
	}

	if len(nonce) != 16 {
		return nil, fmt.Errorf("invalid nonce length: expected 16 characters, got %d", len(nonce))
	}

	return &SIWEMessage{
		Domain:    v.domain,
		Nonce:     nonce,
		Statement: "Login to DappLink",
	}, nil
}

func (v *SIWEVerifier) VerifySignature(message, signature string) (string, error) {
	siweMsg, err := v.ParseSIWEMessage(message)
	if err != nil {
		log.Error("Failed to parse SIWE message", "err", err)
		return "", fmt.Errorf("invalid SIWE message: %w", err)
	}
	if siweMsg.Domain != v.domain {
		return "", fmt.Errorf("domain mismatch: expected %s, got %s", v.domain, siweMsg.Domain)
	}
	messageHash := accounts.TextHash([]byte(message))
	sigBytes, err := hexutil.Decode(signature)
	if err != nil {
		return "", fmt.Errorf("invalid signature format: %w", err)
	}
	if len(sigBytes) != 65 {
		return "", errors.New("signature must be 65 bytes long")
	}
	if sigBytes[64] >= 27 {
		sigBytes[64] -= 27
	}
	pubKey, err := crypto.SigToPub(messageHash, sigBytes)
	if err != nil {
		return "", fmt.Errorf("failed to recover public key: %w", err)
	}
	recoveredAddr := crypto.PubkeyToAddress(*pubKey)
	if siweMsg.Address == "" {
		log.Info("Signature verified successfully (simplified format)", "address", recoveredAddr.Hex())
		return recoveredAddr.Hex(), nil
	}

	address := strings.ToLower(siweMsg.Address)
	if !common.IsHexAddress(address) {
		return "", errors.New("invalid Ethereum address")
	}

	expectedAddr := common.HexToAddress(address)
	if recoveredAddr != expectedAddr {
		log.Warn("Address mismatch", "expected", expectedAddr.Hex(), "recovered", recoveredAddr.Hex())
		return "", errors.New("signature verification failed: address mismatch")
	}
	log.Info("Signature verified successfully", "address", recoveredAddr.Hex())
	return recoveredAddr.Hex(), nil
}

func (v *SIWEVerifier) GenerateJWT(address string, expirationHours int) (string, error) {
	claims := JWTClaims{
		Address: strings.ToLower(address),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expirationHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(v.jwtSecret)
	if err != nil {
		return "", fmt.Errorf("failed to generate JWT: %w", err)
	}
	return tokenString, nil
}

func (v *SIWEVerifier) VerifyJWT(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return v.jwtSecret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse JWT: %w", err)
	}
	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}
