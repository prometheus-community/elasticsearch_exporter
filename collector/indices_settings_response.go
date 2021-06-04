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

package collector

// IndicesSettingsResponse is a representation of Elasticsearch Settings for each Index
type IndicesSettingsResponse map[string]Index

// Index defines the struct of the tree for the settings of each index
type Index struct {
	Settings Settings `json:"settings"`
}

// Settings defines current index settings
type Settings struct {
	IndexInfo IndexInfo `json:"index"`
}

// IndexInfo defines the blocks of the current index
type IndexInfo struct {
	Blocks  Blocks  `json:"blocks"`
	Mapping Mapping `json:"mapping"`
}

// Blocks defines whether current index has read_only_allow_delete enabled
type Blocks struct {
	ReadOnly string `json:"read_only_allow_delete"`
}

// Mapping defines mapping settings
type Mapping struct {
	TotalFields TotalFields `json:"total_fields"`
}

// TotalFields defines the limit on the number of mapped fields
type TotalFields struct {
	Limit string `json:"limit"`
}
