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
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"testing"

	belaur "github.com/bhojpur/belaur"
)

func TestInitCA(t *testing.T) {
	belaur.Cfg = &belaur.Config{}
	tmp, _ := ioutil.TempDir("", "TestInitCA")
	belaur.Cfg.DataPath = tmp

	c, err := InitCA()
	if err != nil {
		t.Fatal(err)
	}

	// Get root CA cert path
	caCertPath, caKeyPath := c.GetCACertPath()

	// Load CA plain
	caPlain, err := tls.LoadX509KeyPair(caCertPath, caKeyPath)
	if err != nil {
		t.Fatal(err)
	}

	// Parse certificate
	ca, err := x509.ParseCertificate(caPlain.Certificate[0])
	if err != nil {
		t.Fatal(err)
	}

	// Create cert pool and load ca root
	certPool := x509.NewCertPool()
	rootCA, err := ioutil.ReadFile(caCertPath)
	if err != nil {
		t.Fatal(err)
	}

	ok := certPool.AppendCertsFromPEM(rootCA)
	if !ok {
		t.Fatalf("Cannot append root cert to cert pool!\n")
	}

	_, err = ca.Verify(x509.VerifyOptions{
		Roots:   certPool,
		DNSName: orgDNS,
	})
	if err != nil {
		t.Fatal(err)
	}

	err = c.CleanupCerts(caCertPath, caKeyPath)
	if err != nil {
		t.Fatal(err)
	}
}

func TestCreateSignedCert(t *testing.T) {
	belaur.Cfg = &belaur.Config{}
	tmp, _ := ioutil.TempDir("", "TestCreateSignedCert")
	belaur.Cfg.DataPath = tmp

	c, err := InitCA()
	if err != nil {
		t.Fatal(err)
	}

	// Get root ca cert path
	caCertPath, caKeyPath := c.GetCACertPath()

	certPath, keyPath, err := c.CreateSignedCert()
	if err != nil {
		t.Fatal(err)
	}

	// Load CA plain
	caPlain, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		t.Fatal(err)
	}

	// Parse certificate
	ca, err := x509.ParseCertificate(caPlain.Certificate[0])
	if err != nil {
		t.Fatal(err)
	}

	// Create cert pool and load ca root
	certPool := x509.NewCertPool()
	rootCA, err := ioutil.ReadFile(caCertPath)
	if err != nil {
		t.Fatal(err)
	}

	ok := certPool.AppendCertsFromPEM(rootCA)
	if !ok {
		t.Fatalf("Cannot append root cert to cert pool!\n")
	}

	_, err = ca.Verify(x509.VerifyOptions{
		Roots:   certPool,
		DNSName: orgDNS,
	})
	if err != nil {
		t.Fatal(err)
	}

	err = c.CleanupCerts(caCertPath, caKeyPath)
	if err != nil {
		t.Fatal(err)
	}
	err = c.CleanupCerts(certPath, keyPath)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGenerateTLSConfig(t *testing.T) {
	belaur.Cfg = &belaur.Config{}
	tmp, _ := ioutil.TempDir("", "TestGenerateTLSConfig")
	belaur.Cfg.DataPath = tmp

	c, err := InitCA()
	if err != nil {
		t.Fatal(err)
	}

	// Get root ca cert path
	caCertPath, caKeyPath := c.GetCACertPath()

	certPath, keyPath, err := c.CreateSignedCert()
	if err != nil {
		t.Fatal(err)
	}

	// Generate TLS Config
	_, err = c.GenerateTLSConfig(certPath, keyPath)
	if err != nil {
		t.Fatal(err)
	}

	err = c.CleanupCerts(caCertPath, caKeyPath)
	if err != nil {
		t.Fatal(err)
	}
	err = c.CleanupCerts(certPath, keyPath)
	if err != nil {
		t.Fatal(err)
	}
}
