package gufu

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"golang.org/x/crypto/pbkdf2"
)

// Gera uma palavra aleatoria de tamanho x
func randomString(x int) string {
	rand.Seed(time.Now().UnixNano())
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	s := make([]rune, x)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}

// Gera uma chave derivada usando o algoritmo PBKDF2 (Password-Based Key Derivation Function 2)
func generateKeyEncryptation(salt, passphrase string, keySize, iterationCount int) ([]byte, error) {
	saltBytes, err := hex.DecodeString(salt)
	if err != nil {
		return nil, err
	}
	key := pbkdf2.Key([]byte(passphrase), saltBytes, iterationCount, keySize, sha1.New)
	return key, nil
}

// Criptografa o texto simples fornecido usando AES no modo CBC
func encryptAES(salt, iv, passphrase, plaintext string, keySize, iterationCount int) (string, error) {
	key, err := generateKeyEncryptation(salt, passphrase, keySize, iterationCount)
	if err != nil {
		return "", err
	}
	ivBytes, err := hex.DecodeString(iv)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	plaintextBytes := []byte(plaintext)
	padding := aes.BlockSize - len(plaintextBytes)%aes.BlockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	plaintextBytes = append(plaintextBytes, padtext...)

	ciphertext := make([]byte, aes.BlockSize+len(plaintextBytes))
	copy(ciphertext[:aes.BlockSize], ivBytes)

	mode := cipher.NewCBCEncrypter(block, ivBytes)
	mode.CryptBlocks(ciphertext[aes.BlockSize:], plaintextBytes)

	return base64.StdEncoding.EncodeToString(ciphertext[aes.BlockSize:]), nil
}

// Criptografa o texto usando a senha
func encryptText(text, passphrase string) (string, error) {
	const (
		salt      = "3FF2EC019C627B945225DEBAD71A01B6985FE84C95A70EB132882F88C0A59A55"
		iv        = "F27D5C9927726BCEFE7510B1BDD3D137"
		keySize   = 16
		iterCount = 10
	)
	return encryptAES(salt, iv, passphrase, text, keySize, iterCount)
}

// Cria o JSON da requests com os dados encriptados
func criptografarRequestParams(req string) (string, error) {
	if req == "" {
		return "", errors.New("requestParams está vazio")
	}
	p := randomString(25)
	encryptedText, err := encryptText(req, p)
	if err != nil {
		return "", err
	}
	decodedSuffix, err := base64.StdEncoding.DecodeString("WWNrbjlTQUZwcU04SzlCMnVKWWVlSFpqRkho")
	if err != nil {
		return "", err
	}
	er := encryptedText + string(decodedSuffix) + p
	return fmt.Sprintf(`{"requestParams":"%v"}`, er), nil
}
