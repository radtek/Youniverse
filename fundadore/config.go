package fundadore

type resourceExecInfo struct {
	Delay           int    `json:"delay"`
	ModeServer      bool   `json:"modeServer"`
	ContinueOnError bool   `json:"continueOnError"`
	PEType          string `json:"peType"`
	PEEntry         string `json:"peEntry"`
	FileType        string `json:"fileType"`
	WorkPath        string `json:"workPath"`
	ShowMode        int    `json:"showMode"`
	Parameter       string `json:"parameter"`
}

type resourceSaveInfo struct {
	Must   bool   `json:"must"`
	Type   string `json:"type"`
	OsType string `json:"os_type"`

	Path  string `json:"path"`
	Param string `json:"param"`
}

type TaskOld struct {
	Name string           `json:"name"`
	Hash string           `json:"hash"`
	Save resourceSaveInfo `json:"save"` //Exec     execInfo `json:"save"`
}

type Task struct {
	Name     string `json:"name"`
	Hash     string `json:"hash"`
	SavePath string `json:"savePath"`
	Exec     string `json:"save"` //Exec     execInfo `json:"save"`
}

type Settings struct {
	TasksURL string `json:"tasks_url"`
}
