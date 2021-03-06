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
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"io/ioutil"
	"math/big"
	"os"
	"path/filepath"
	"time"

	belaur "github.com/bhojpur/belaur"
)

const (
	rsaBits      = 2048
	maxValidCA   = 175200 // 20 years
	maxValidCERT = 48     // 48 hours
	orgName      = "Bhojpur Consulting"
	orgDNS       = "bhojpur.net"

	// CA key name
	certName = "ca.crt"
	keyName  = "ca.key"

	// Hostname used by go-plugin
	goPluginHostname = "unused"
)

var (
	// errCertNotAppended is thrown when the root CA cert cannot be appended to the pool.
	errCertNotAppended = errors.New("cannot append root CA cert to cert pool")
)

// CA represents one generated CA.
type CA struct {
	caCertPath string
	caKeyPath  string
}

// CAAPI represents the interface used to handle certificates.
type CAAPI interface {
	// CreateSignedCert creates a new signed certificate.
	// First return param is the public cert.
	// Second return param is the private key.
	CreateSignedCert() (string, string, error)

	// CreateSignedCertWithValidOpts create a new signed certificate
	// with the given options.
	// First return param is the public cert.
	// Second return param is the private key.
	CreateSignedCertWithValidOpts(hostname string, hoursBeforeValid, hoursAfterValid time.Duration) (string, string, error)

	// GenerateTLSConfig generates a TLS config.
	// It requires the path to the cert and the key.
	GenerateTLSConfig(certPath, keyPath string) (*tls.Config, error)

	// CleanupCerts cleans up the certs at the given path.
	CleanupCerts(crt, key string) error

	// GetCACertPath returns the public cert and private key
	// of the CA.
	GetCACertPath() (string, string)
}

// InitCA setups a new instance of CA and generates a new CA if not already exists.
func InitCA() (*CA, error) {
	t := &CA{
		caCertPath: filepath.Join(belaur.Cfg.CAPath, certName),
		caKeyPath:  filepath.Join(belaur.Cfg.CAPath, keyName),
	}
	return t, t.generateCA()
}

// generateCA generates the CA and puts the certs into the data folder.
// If they are already existing, nothing will be done.
func (c *CA) generateCA() error {
	// Check if they are already existing
	_, certErr := os.Stat(c.caCertPath)
	_, keyErr := os.Stat(c.caKeyPath)

	// Both exist, skip
	if certErr == nil && keyErr == nil {
		return nil
	}

	// Set time range for cert validation
	notBefore := time.Now()
	notAfter := notBefore.Add(time.Hour * maxValidCA)

	// Generate serial number
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return err
	}

	// Generate CA template
	template := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{orgName},
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		IsCA:                  true,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{orgDNS, goPluginHostname},
	}

	// Generate the key
	key, err := rsa.GenerateKey(rand.Reader, rsaBits)
	if err != nil {
		return err
	}

	// Create certificate authority
	derBytes, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		return err
	}

	// Write out the ca.crt file
	certOut, err := os.Create(c.caCertPath)
	if err != nil {
		return err
	}
	err = pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	if err != nil {
		return err
	}
	err = certOut.Close()
	if err != nil {
		return err
	}

	// Write out the ca.key file
	keyOut, err := os.OpenFile(c.caKeyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}

	// Write out key file in PKCS#8 format
	privateKey, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return err
	}
	err = pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privateKey})
	if err != nil {
		return err
	}
	err = keyOut.Close()
	if err != nil {
		return err
	}

	return nil
}

// CreateSignedCertWithValidOpts creates a signed certificate by the CA.
// It accepts hoursBeforeValid and hoursAfterValid.
func (c *CA) CreateSignedCertWithValidOpts(hostname string, hoursBeforeValid, hoursAfterValid time.Duration) (string, string, error) {
	// Load CA plain
	caPlain, err := tls.LoadX509KeyPair(c.caCertPath, c.caKeyPath)
	if err != nil {
		return "", "", err
	}

	// Parse certificate
	ca, err := x509.ParseCertificate(caPlain.Certificate[0])
	if err != nil {
		return "", "", err
	}

	// Generate serial number
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return "", "", err
	}

	// Set time range for cert validation
	notBefore := time.Now().Add(time.Hour * (-1 * hoursBeforeValid))
	notAfter := time.Now().Add(time.Hour * hoursAfterValid)

	// Prepare certificate
	cert := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{orgName},
		},
		NotBefore:    notBefore,
		NotAfter:     notAfter,
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
		DNSNames:     []string{orgDNS, goPluginHostname},
	}

	// Add additional hostname if provided
	if hostname != "" {
		cert.DNSNames = append(cert.DNSNames, hostname)
	}

	priv, _ := rsa.GenerateKey(rand.Reader, rsaBits)
	pub := &priv.PublicKey

	// Sign the certificate
	certSigned, err := x509.CreateCertificate(rand.Reader, cert, ca, pub, caPlain.PrivateKey)
	if err != nil {
		return "", "", err
	}

	// Public key
	certOut, err := ioutil.TempFile("", "crt")
	if err != nil {
		return "", "", err
	}

	err = pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: certSigned})
	if err != nil {
		return "", "", err
	}
	err = certOut.Close()
	if err != nil {
		return "", "", err
	}

	// Private key
	keyOut, err := ioutil.TempFile("", "key")
	if err != nil {
		return "", "", err
	}

	// Write out key file in PKCS#8 format
	privateKey, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return "", "", err
	}

	err = pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privateKey})
	if err != nil {
		return "", "", err
	}
	err = keyOut.Close()
	if err != nil {
		return "", "", err
	}

	return certOut.Name(), keyOut.Name(), nil
}

// CreateSignedCert creates a new key pair which is signed by the CA.
func (c *CA) CreateSignedCert() (string, string, error) {
	return c.CreateSignedCertWithValidOpts("", 1, maxValidCERT)
}

// GenerateTLSConfig generates a new TLS config based on given
// certificate path and key path.
func (c *CA) GenerateTLSConfig(certPath, keyPath string) (*tls.Config, error) {
	// Load certificate
	certificate, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return nil, err
	}

	// Create certificate pool
	certPool := x509.NewCertPool()
	caCert, err := ioutil.ReadFile(c.caCertPath)
	if err != nil {
		return nil, err
	}

	// Append cert to cert pool
	ok := certPool.AppendCertsFromPEM(caCert)
	if !ok {
		return nil, errCertNotAppended
	}

	return &tls.Config{
		Certificates: []tls.Certificate{certificate},
		ClientCAs:    certPool,
		RootCAs:      certPool,
	}, nil
}

// CleanupCerts removes certificates at the given path.
func (c *CA) CleanupCerts(crt, key string) error {
	if err := os.Remove(crt); err != nil {
		return err
	}
	return os.Remove(key)
}

// GetCACertPath returns the path to the cert and key from the root CA.
func (c *CA) GetCACertPath() (string, string) {
	return c.caCertPath, c.caKeyPath
}
