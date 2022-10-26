package crypt

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"github.com/andreburgaud/crypt2go/ecb"
	"github.com/andreburgaud/crypt2go/padding"
	"log"
	//
	//"github.com/andreburgaud/crypt2go/ecb"
	//"github.com/andreburgaud/crypt2go/padding"
)

//aes-cbc-PKCS7 加密
func EncryptAESCBCPKCS7(data []byte, key []byte, iv []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	data = PKCS7Padding(data)

	cipherData := make([]byte, len(data))
	cbcEncrypt := cipher.NewCBCEncrypter(block, iv)
	cbcEncrypt.CryptBlocks(cipherData, data)
	return cipherData, err
}

//aes-cbc-PKCS7 解密
func DecryptAESCBCPKCS7(cipherData []byte, key []byte, iv []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	cbcDecrypt := cipher.NewCBCDecrypter(block, iv)
	data := make([]byte, len(cipherData))
	cbcDecrypt.CryptBlocks(data, cipherData)
	return PKCS7UnPadding(data), nil
}

func PKCS7Padding(ciphertext []byte) []byte {
	padding := aes.BlockSize - len(ciphertext)%aes.BlockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}
func PKCS7UnPadding(plantText []byte) []byte {
	length := len(plantText)
	log.Printf("PKCS7UnPadding()-%d", length)
	unpadding := int(plantText[length-1])
	return plantText[:(length - unpadding)]
}

//aes-ecb-PKCS7 加密
func EncryptAESECBPKCS7(data []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	mode := ecb.NewECBEncrypter(block)
	//dataPad := PKCS7Padding(data)
	dataPad, err := padding.NewPkcs7Padding(mode.BlockSize()).Pad(data) // pad last block of plaintext if block size less than block cipher size
	if err != nil {
		return nil, err
	}
	ct := make([]byte, len(dataPad))
	mode.CryptBlocks(ct, dataPad)
	return ct, nil
}
