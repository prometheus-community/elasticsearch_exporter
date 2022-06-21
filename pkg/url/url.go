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
	"errors"
	"fmt"
	"net/url"
	"strings"
)

const defaultEsURI = "http://localhost:9200"

func GetEsURL(esURI, cloudID string) (*url.URL, error) {
	var uri string
	var err error
	if len(esURI) == 0 && len(cloudID) == 0 {
		uri = defaultEsURI
	} else {
		if len(esURI) > 0 && len(cloudID) > 0 {
			return nil, errors.New("cannot create client: both es.uri and ES_CLOUD_ID are set")
		}

		if len(esURI) > 0 {
			uri = esURI
		}

		if len(cloudID) > 0 {
			uri, err = addrFromCloudID(cloudID)
			if err != nil {
				return nil, err
			}
		}
	}

	esURL, err := url.Parse(uri)
	if err != nil {
		return nil, fmt.Errorf("failed to parse url: %v", err)
	}

	return esURL, nil
}

// addrFromCloudID extracts the Elasticsearch URL from CloudID.
// See: https://www.elastic.co/guide/en/cloud/current/ec-cloud-id.html
//
// This function was originally copied from https://github.com/elastic/go-elasticsearch/blob/8134a159aafedf58af2780ebb3a30ec1938956f3/elasticsearch.go#L365-L383
func addrFromCloudID(input string) (string, error) {
	var scheme = "https://"

	values := strings.Split(input, ":")
	if len(values) != 2 {
		return "", fmt.Errorf("unexpected format: %q", input)
	}
	data, err := base64.StdEncoding.DecodeString(values[1])
	if err != nil {
		return "", err
	}
	parts := strings.Split(string(data), "$")

	if len(parts) < 2 {
		return "", fmt.Errorf("invalid encoded value: %s", parts)
	}

	return fmt.Sprintf("%s%s.%s", scheme, parts[1], parts[0]), nil
}
