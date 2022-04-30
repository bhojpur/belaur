package security

// Copyright (c) 2018 Bhojpur Consulting Private Limited, India. All rights reserved.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"

	belaur "github.com/bhojpur/belaur"
)

const (
	secretCheckKey   = "BELAUR_CHECK_SECRET"
	secretCheckValue = "!CHECK_ME!"
)

func (v *Vault) legacyDecrypt(data []byte) ([]byte, error) {
	if len(data) < 1 {
		belaur.Cfg.Logger.Info("the vault is empty")
		return []byte{}, nil
	}
	paddedPassword := v.pad(v.cert)
	ci := base64.URLEncoding.EncodeToString(paddedPassword)
	block, err := aes.NewCipher([]byte(ci[:aes.BlockSize]))
	if err != nil {
		return []byte{}, err
	}

	decodedMsg, err := base64.URLEncoding.DecodeString(string(data))
	if err != nil {
		return []byte{}, err
	}

	if (len(decodedMsg) % aes.BlockSize) != 0 {
		return []byte{}, errors.New("blocksize must be multiple of decoded message length")
	}

	iv := decodedMsg[:aes.BlockSize]
	msg := decodedMsg[aes.BlockSize:]

	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(msg, msg)

	unpadMsg, err := v.unpad(msg)
	if err != nil {
		return []byte{}, err
	}

	if !bytes.Contains(unpadMsg, []byte(secretCheckValue)) {
		return []byte{}, errors.New("possible mistyped password")
	}
	return unpadMsg, nil
}

// Pad pads the src with 0x04 until block length.
func (v *Vault) pad(src []byte) []byte {
	padding := aes.BlockSize - len(src)%aes.BlockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(src, padtext...)
}

// Unpad removes the padding from pad.
func (v *Vault) unpad(src []byte) ([]byte, error) {
	length := len(src)
	unpadding := int(src[length-1])

	if unpadding > length {
		return nil, errors.New("possible mistyped password")
	}

	return src[:(length - unpadding)], nil
}
