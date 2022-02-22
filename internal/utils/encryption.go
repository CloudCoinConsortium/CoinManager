package utils

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/logger"
)


func Encrypt(ctx context.Context, an string, sn uint32, rIdx int, body []byte) ([]byte, error) {
  key, _ := hex.DecodeString(an)

  logger.L(ctx).Debugf("RAIDA%d Encrypting body with key %v body %v, AES key size %d", rIdx, key, body, aes.BlockSize)

  block, err := aes.NewCipher(key)
  if err != nil {
    logger.L(ctx).Errorf("Failed to create cipher: %s", err.Error())
    return nil, err
  }

  nonce := GetNonce(ctx, sn)
  cipherText := make([]byte, len(body))

  encryptStream := cipher.NewCTR(block, nonce)
  encryptStream.XORKeyStream(cipherText, body)

  logger.L(ctx).Debugf("Encrypted body %v", cipherText)

  return cipherText, nil 
}

func Decrypt(ctx context.Context, an string, sn uint32, rIdx int, body []byte) ([]byte, error) {
  key, _ := hex.DecodeString(an)

  logger.L(ctx).Debugf("Decrypting body with key %v body %v, AES key size %d", key, body, aes.BlockSize)

  block, err := aes.NewCipher(key)
  if err != nil {
    logger.L(ctx).Errorf("Failed to create cipher: %s", err.Error())
    return nil, err
  }

  nonce := GetNonce(ctx, sn)
  plainText := make([]byte, len(body))

  encryptStream := cipher.NewCTR(block, nonce)
  encryptStream.XORKeyStream(plainText, body)

  logger.L(ctx).Debugf("Decrypted body %v", plainText)

  return plainText, nil 
}

func GetNonce(ctx context.Context, sn uint32) []byte {
  nonce := make([]byte, 16)
  nonce[0] = 0
  nonce[1] = 0
  nonce[2] = 0
  nonce[3] = 0x11
  nonce[4] = 0x11
  nonce[5] = byte((sn >> 16) & 0xff)
  nonce[6] = byte((sn >> 8) & 0xff)
  nonce[7] = byte((sn) & 0xff)

  logger.L(ctx).Debugf("nonce %v", nonce)

  return nonce
}
