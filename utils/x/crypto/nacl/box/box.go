// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package box authenticates and encrypts messages using public-key cryptography.

Box uses Curve25519, XSalsa20 and Poly1305 to encrypt and authenticate
messages. The length of messages is not hidden.

It is the caller's responsibility to ensure the uniqueness of nonces—for
example, by using nonce 1 for the first message, nonce 2 for the second
message, etc. Nonces are long enough that randomly generated nonces have
negligible risk of collision.

This package is interoperable with NaCl: https://nacl.cr.yp.to/box.html.
*/
package box // import "github.com/GodSlave/mqant/utils/x/crypto/nacl/box"

import (
	"io"

	"github.com/GodSlave/mqant/utils/x/crypto/curve25519"
	"github.com/GodSlave/mqant/utils/x/crypto/nacl/secretbox"
	"github.com/GodSlave/mqant/utils/x/crypto/salsa20/salsa"
)

// Overhead is the number of bytes of overhead when boxing a message.
const Overhead = secretbox.Overhead

// GenerateKey generates a new public/private key pair suitable for use with
// Seal and Open.
func GenerateKey(rand io.Reader) (publicKey, privateKey *[32]byte, err error) {
	publicKey = new([32]byte)
	privateKey = new([32]byte)
	_, err = io.ReadFull(rand, privateKey[:])
	if err != nil {
		publicKey = nil
		privateKey = nil
		return
	}

	curve25519.ScalarBaseMult(publicKey, privateKey)
	return
}

var zeros [16]byte

// Precompute calculates the shared key between peersPublicKey and privateKey
// and writes it to sharedKey. The shared key can be used with
// OpenAfterPrecomputation and SealAfterPrecomputation to speed up processing
// when using the same pair of keys repeatedly.
func Precompute(sharedKey, peersPublicKey, privateKey *[32]byte) {
	curve25519.ScalarMult(sharedKey, privateKey, peersPublicKey)
	salsa.HSalsa20(sharedKey, &zeros, sharedKey, &salsa.Sigma)
}

// Seal appends an encrypted and authenticated copy of message to out, which
// will be Overhead bytes longer than the original and must not overlap. The
// nonce must be unique for each distinct message for a given pair of keys.
func Seal(out, message []byte, nonce *[24]byte, peersPublicKey, privateKey *[32]byte) []byte {
	var sharedKey [32]byte
	Precompute(&sharedKey, peersPublicKey, privateKey)
	return secretbox.Seal(out, message, nonce, &sharedKey)
}

// SealAfterPrecomputation performs the same actions as Seal, but takes a
// shared key as generated by Precompute.
func SealAfterPrecomputation(out, message []byte, nonce *[24]byte, sharedKey *[32]byte) []byte {
	return secretbox.Seal(out, message, nonce, sharedKey)
}

// Open authenticates and decrypts a box produced by Seal and appends the
// message to out, which must not overlap box. The output will be Overhead
// bytes smaller than box.
func Open(out, box []byte, nonce *[24]byte, peersPublicKey, privateKey *[32]byte) ([]byte, bool) {
	var sharedKey [32]byte
	Precompute(&sharedKey, peersPublicKey, privateKey)
	return secretbox.Open(out, box, nonce, &sharedKey)
}

// OpenAfterPrecomputation performs the same actions as Open, but takes a
// shared key as generated by Precompute.
func OpenAfterPrecomputation(out, box []byte, nonce *[24]byte, sharedKey *[32]byte) ([]byte, bool) {
	return secretbox.Open(out, box, nonce, sharedKey)
}
