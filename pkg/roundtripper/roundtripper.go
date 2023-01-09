// Copyright 2022 The Prometheus Authors
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

package roundtripper

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

const (
	service = "es"
)

type AWSSigningTransport struct {
	t      http.RoundTripper
	creds  aws.Credentials
	region string
	log    log.Logger
}

func NewAWSSigningTransport(transport http.RoundTripper, region string, log log.Logger) (*AWSSigningTransport, error) {
	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(region))
	if err != nil {
		_ = level.Error(log).Log("msg", "fail to load aws default config", "err", err)
		return nil, err
	}

	creds, err := cfg.Credentials.Retrieve(context.Background())
	if err != nil {
		_ = level.Error(log).Log("msg", "fail to retrive aws credentials", "err", err)
		return nil, err
	}

	return &AWSSigningTransport{
		t:      transport,
		region: region,
		creds:  creds,
		log:    log,
	}, err
}

func (a *AWSSigningTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	signer := v4.NewSigner()
	payloadHash, newReader, err := hashPayload(req.Body)
	if err != nil {
		_ = level.Error(a.log).Log("msg", "fail to hash request body", "err", err)
		return nil, err
	}
	req.Body = newReader
	err = signer.SignHTTP(context.Background(), a.creds, req, payloadHash, service, a.region, time.Now())
	if err != nil {
		_ = level.Error(a.log).Log("msg", "fail to sign request body", "err", err)
		return nil, err
	}
	return a.t.RoundTrip(req)
}

func hashPayload(r io.ReadCloser) (string, io.ReadCloser, error) {
	var newReader io.ReadCloser
	payload := []byte("")
	if r != nil {
		defer r.Close()
		payload, err := ioutil.ReadAll(r)
		if err != nil {
			return "", newReader, err
		}
		newReader = ioutil.NopCloser(bytes.NewReader(payload))
	}
	hash := sha256.Sum256(payload)
	payloadHash := hex.EncodeToString(hash[:])
	return payloadHash, newReader, nil
}
