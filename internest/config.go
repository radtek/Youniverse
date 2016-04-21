package internest

type saveInfo struct {
	Type string `json:"type"`
	Path string `json:"path"`
}

type saveInfos struct {
	X86     saveInfo `json:"x86"`
	X64     saveInfo `json:"x64"`
	Arm     saveInfo `json:"arm"`
}

type Resource struct {
	Name string    `json:"name"`
	Hash string    `json:"hash"`
	Save saveInfos `json:"save"`
}

type Settings struct {
	Resources []Resource `json:"resources"`
}
