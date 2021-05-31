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

// IndicesMappingsResponse is a representation of elasticsearch mappings for each index
type IndicesMappingsResponse map[string]IndexMapping

// IndexMapping defines the struct of the tree for the mappings of each index
type IndexMapping struct {
	Mappings IndexMappings `json:"mappings"`
}

// IndexMappings defines all index mappings
type IndexMappings struct {
	Properties IndexMappingProperties `json:"properties"`
}

// IndexMappingProperties defines all the properties of the current mapping
type IndexMappingProperties map[string]*IndexMappingProperty

// IndexMappingFields defines all the fields of the current mapping
type IndexMappingFields map[string]*IndexMappingField

// IndexMappingProperty defines a single property of the current index properties
type IndexMappingProperty struct {
	Type       *string                `json:"type"`
	Properties IndexMappingProperties `json:"properties"`
	Fields     IndexMappingFields     `json:"fields"`
}

// IndexMappingField defines a single property of the current index field
type IndexMappingField struct {
	Type       *string                `json:"type"`
	Properties IndexMappingProperties `json:"properties"`
	Fields     IndexMappingFields     `json:"fields"`
}
