// Copyright 2021 The Prometheus Authors
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
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws/credentials"
	v4 "github.com/aws/aws-sdk-go/aws/signer/v4"
)

const (
	service = "es"
)

type AWSSigningTransport struct {
	DefaultTransport *http.Transport
	Credentials      *credentials.Credentials
	Region           string
}

// RoundTrip implementation
func (a AWSSigningTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	signer := v4.NewSigner(a.Credentials)

	// body is nil as we never send data to Elastic, just get
	if _, err := signer.Sign(req, nil, service, a.Region, time.Now()); err != nil {
		return nil, err
	}

	return a.DefaultTransport.RoundTrip(req)
}
