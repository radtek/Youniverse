package fundadore

type saveInfo struct {
	Must   bool   `json:"must"`
	Type   string `json:"type"`
	OsType string `json:"os_type"`

	Path  string `json:"path"`
	Param string `json:"param"`
}

type Task struct {
	Name string   `json:"name"`
	Hash string   `json:"hash"`
	Save saveInfo `json:"save"`
}

type Settings struct {
	TasksURL    string `json:"tasks_url"`
}

