package main

import (
	"bytes"
	"crypto/cipher"
	"crypto/des"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
)

func md5_(s string) string {
	m := md5.New()
	m.Write([]byte(s))
	return hex.EncodeToString(m.Sum(nil))[8:24]
}

//Des对称加密

func DesEncrypt(origData_, key_ string) (string, error) {
	origData := []byte(origData_)
	key := []byte(key_)
	block, err := des.NewCipher(key)
	if err != nil {
		return "", err
	}
	origData = PKCS5Padding(origData, block.BlockSize())
	blockMode := cipher.NewCBCEncrypter(block, key)
	crypted := make([]byte, len(origData))
	blockMode.CryptBlocks(crypted, origData)
	tmp := base64.StdEncoding.EncodeToString(crypted)
	return tmp, nil
}
func PKCS5Padding(cipherText []byte, blockSize int) []byte {
	padding := blockSize - len(cipherText)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(cipherText, padText...)
}

//Des对称解密
func DesDecrypt(crypted_, key_ string) (string, error) {
	crypted, err := base64.StdEncoding.DecodeString(crypted_)
	key := []byte(key_)
	block, err := des.NewCipher(key)
	if err != nil {
		return "", err
	}
	blockMode := cipher.NewCBCDecrypter(block, key)
	origData := make([]byte, len(crypted))
	// origData := crypted
	blockMode.CryptBlocks(origData, crypted)
	origData = PKCS5UnPadding(origData)
	// origData = ZeroUnPadding(origData)
	return string(origData), nil
}
func PKCS5UnPadding(origData []byte) []byte {
	length := len(origData)
	// 去掉最后一个字节 unpadding 次
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}
