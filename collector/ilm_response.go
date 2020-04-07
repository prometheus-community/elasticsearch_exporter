package collector

type IlmResponse struct {
	Indices map[string]IlmIndexResponse `json:"indices"`
}

type IlmIndexResponse struct {
	Index   string `json:"index"`
	Managed bool   `json:"managed"`
	Phase   string `json:"phase"`
	Action  string `json:"action"`
	Step    string `json:"step"`
}
