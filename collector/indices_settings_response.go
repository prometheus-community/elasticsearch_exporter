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
	Blocks Blocks `json:"blocks"`
}

// Blocks defines whether current index has read_only_allow_delete enabled
type Blocks struct {
	ReadOnly string `json:"read_only_allow_delete"`
}
