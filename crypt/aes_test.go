package crypt

import (
	"encoding/hex"
	"log"
	"testing"
)

func TestAESCBCPKCS7(t *testing.T) {
	data := "1234567890"
	key := "0CoJUm6Qyw8W8jud"
	iv := "0102030405060708"
	encryptedData, err := EncryptAESCBCPKCS7([]byte(data), []byte(key), []byte(iv))
	if err != nil {
		t.Fatal(err)
	}

	log.Println("encrypt " + hex.EncodeToString(encryptedData))
	dataDe, err := DecryptAESCBCPKCS7(encryptedData, []byte(key), []byte(iv))
	if err != nil {
		t.Fatal(err)
	}
	log.Println(string(dataDe))

	log.Println(string([]byte(data)))

}
