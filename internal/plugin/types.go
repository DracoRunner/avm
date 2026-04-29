package plugin

type Manifest struct {
	Name         string `json:"name"`
	Version      string `json:"version"`
	APIVersion   int    `json:"api_version"`
	Description  string `json:"description"`
	SectionLabel string `json:"section_label"`
	Homepage     string `json:"homepage"`
}

type AliasDetail struct {
	Command     string `json:"command"`
	Description string `json:"description"`
	Source      string `json:"source"`
}

type ExportResponse struct {
	APIVersion int                    `json:"api_version"`
	Aliases    map[string]interface{} `json:"aliases"` // interface{} to support both string and AliasDetail
}

type ResolvedAlias struct {
	Command     string
	Description string
	PluginName  string
	SectionName string
}
