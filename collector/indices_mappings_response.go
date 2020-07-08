package collector

// IndicesMappingsResponse is a representation of elasticsearch mappings for each index
type IndicesMappingsResponse map[string]IndexMapping

// IndexMapping defines the struct of the tree for the mappings of each index
type IndexMapping struct {
	Mappings Mappings `json:"mappings"`
}

// Mappings defines all index mappings
type Mappings struct {
	Properties Properties `json:"properties"`
}

// Properties defines all the properties of the current mapping
type Properties map[string]*Property

// Fields defines all the fields of the current mapping
type Fields map[string]*Field

// Property defines a single property of the current index properties
type Property struct {
	Type *string `json:"type"`
	Properties Properties `json:"properties"`
	Fields Fields `json:"fields"`
}

// Field defines a single property of the current index field
type Field struct {
	Type *string `json:"type"`
	Properties Properties `json:"properties"`
	Fields Fields `json:"fields"`
}