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

package url

import (
	"encoding/base64"
	"fmt"
	"regexp"
	"testing"
)

func TestAddrFromCloudID(t *testing.T) {
	t.Run("Parse", func(t *testing.T) {
		var testdata = []struct {
			in  string
			out string
		}{
			{
				in:  "name:" + base64.StdEncoding.EncodeToString([]byte("host$es_uuid$kibana_uuid")),
				out: "https://es_uuid.host",
			},
			{
				in:  "name:" + base64.StdEncoding.EncodeToString([]byte("host:9243$es_uuid$kibana_uuid")),
				out: "https://es_uuid.host:9243",
			},
			{
				in:  "name:" + base64.StdEncoding.EncodeToString([]byte("host$es_uuid$")),
				out: "https://es_uuid.host",
			},
			{
				in:  "name:" + base64.StdEncoding.EncodeToString([]byte("host$es_uuid")),
				out: "https://es_uuid.host",
			},
		}

		for _, tt := range testdata {
			actual, err := addrFromCloudID(tt.in)
			if err != nil {
				t.Errorf("Unexpected error: %s", err)
			}
			if actual != tt.out {
				t.Errorf("Unexpected output, want=%q, got=%q", tt.out, actual)
			}
		}

	})

	t.Run("Invalid format", func(t *testing.T) {
		input := "foobar"
		_, err := addrFromCloudID(input)
		if err == nil {
			t.Errorf("Expected error for input %q, got %v", input, err)
		}
		match, _ := regexp.MatchString("unexpected format", err.Error())
		if !match {
			t.Errorf("Unexpected error string: %s", err)
		}
	})

	t.Run("Invalid base64 value", func(t *testing.T) {
		input := "foobar:xxxxx"
		_, err := addrFromCloudID(input)
		if err == nil {
			t.Errorf("Expected error for input %q, got %v", input, err)
		}
		match, _ := regexp.MatchString("illegal base64 data", err.Error())
		if !match {
			t.Errorf("Unexpected error string: %s", err)
		}
	})
}

func TestGetEsURL(t *testing.T) {
	cases := map[string]struct {
		esURI       string
		cloudID     string
		expectedUrl string
		expectErr   bool
	}{
		"success - url = defaultURI": {
			expectedUrl: defaultEsURI,
		},
		"success - url = esURI": {
			esURI:       "http://example.com:9200",
			expectedUrl: "http://example.com:9200",
		},
		"success - url = parsed from cloudID": {
			cloudID:     "name:" + base64.StdEncoding.EncodeToString([]byte("host$es_uuid$kibana_uuid")),
			expectedUrl: "https://es_uuid.host",
		},
		"error - both esURI and cloudID are specified": {
			esURI:     "http://example.com:9200",
			cloudID:   "name:" + base64.StdEncoding.EncodeToString([]byte("host$es_uuid$kibana_uuid")),
			expectErr: true,
		},
	}

	for name, tc := range cases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got, err := GetEsURL(tc.esURI, tc.cloudID)

			if err != nil {
				if !tc.expectErr {
					t.Fatalf("Failed to prepare URL: %s", err)
				}
				return
			}

			url := fmt.Sprintf("%s://%s", got.Scheme, got.Host)
			if url != tc.expectedUrl {
				t.Fatalf("Fatiled to parse URL: got:%s, want: %s", url, tc.expectedUrl)
			}
		})
	}
}
