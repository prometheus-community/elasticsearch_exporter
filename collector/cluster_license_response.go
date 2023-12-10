// Copyright 2023 The Prometheus Authors
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

package collector

import "time"

type clusterLicenseResponse struct {
	License struct {
		Status             string    `json:"status"`
		UID                string    `json:"uid"`
		Type               string    `json:"type"`
		IssueDate          time.Time `json:"issue_date"`
		IssueDateInMillis  int64     `json:"issue_date_in_millis"`
		ExpiryDate         time.Time `json:"expiry_date"`
		ExpiryDateInMillis int64     `json:"expiry_date_in_millis"`
		MaxNodes           int       `json:"max_nodes"`
		IssuedTo           string    `json:"issued_to"`
		Issuer             string    `json:"issuer"`
		StartDateInMillis  int64     `json:"start_date_in_millis"`
	} `json:"license"`
}
