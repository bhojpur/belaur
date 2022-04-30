package filehelper

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
	"crypto/sha256"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestGetSHA256Sum(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestGetSHA256Sum")
	sumText := []byte("hello world\n")
	filePath := filepath.Join(tmp, "test.file")
	err := ioutil.WriteFile(filePath, sumText, 0777)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp)
	calcSha, err := GetSHA256Sum(filePath)
	if err != nil {
		t.Fatal(err)
	}
	h := sha256.New()
	if _, err := h.Write(sumText); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(h.Sum(nil), calcSha) {
		t.Fatal("bytes are not identical")
	}
}

func TestCopyFileContents(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestCopyFileContents")
	sumText := []byte("hello world\n")
	filePath := filepath.Join(tmp, "test.file")
	err := ioutil.WriteFile(filePath, sumText, 0777)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp)
	calcSha, err := GetSHA256Sum(filePath)
	if err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(tmp, "copy.file")
	if err := CopyFileContents(filePath, output); err != nil {
		t.Fatal(err)
	}
	copySha, err := GetSHA256Sum(output)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(copySha, calcSha) {
		t.Fatal("bytes are not identical")
	}
}
