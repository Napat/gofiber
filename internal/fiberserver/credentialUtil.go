package fiberserver

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"log"
)

var mykey string = "key001"

var jwtCreds = map[string]*rsa.PrivateKey{}

func jwtCredGet(key string) (*rsa.PrivateKey, error) {
	v, ok := jwtCreds[mykey]
	if !ok {
		return nil, fmt.Errorf("jwtCreds error! key not found: %v\n", key)
	}

	return v, nil
}

func jwtCredSet(key string, v *rsa.PrivateKey) {
	jwtCreds[key] = v
}

// createRsaKey generate a new private/public key pair on each run.
// For demo only, DO NOT DO THIS IN PRODUCTION!
// public/private key should read from secure storage.
func createRsaKey() (*rsa.PrivateKey, error) {
	rng := rand.Reader
	var err error
	privateKey, err := rsa.GenerateKey(rng, 2048)
	if err != nil {
		log.Fatalf("rsa.GenerateKey: %v", err)
		return nil, err
	}

	return privateKey, nil
}
