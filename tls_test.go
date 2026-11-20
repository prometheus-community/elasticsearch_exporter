// Copyright The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func writeTestCA(t *testing.T, path string) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	tmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "test-ca"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		IsCA:                  true,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageCertSign,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestLoadCertificatesFrom(t *testing.T) {
	dir := t.TempDir()

	valid := filepath.Join(dir, "ca.pem")
	writeTestCA(t, valid)
	pool, err := loadCertificatesFrom(valid)
	if err != nil {
		t.Fatalf("valid CA file: unexpected error: %v", err)
	}
	if pool == nil {
		t.Fatal("valid CA file: got nil pool")
	}

	for name, content := range map[string]string{
		"empty":   "",
		"not-pem": "this is not a certificate\n",
		"bad-pem": "-----BEGIN CERTIFICATE-----\nZm9v\n-----END CERTIFICATE-----\n",
	} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(dir, name+".pem")
			if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
				t.Fatal(err)
			}
			if _, err := loadCertificatesFrom(path); err == nil {
				t.Fatalf("expected an error for a CA file with no PEM certificates")
			}
		})
	}

	if _, err := loadCertificatesFrom(filepath.Join(dir, "missing.pem")); err == nil {
		t.Fatal("expected an error for a missing CA file")
	}
}
