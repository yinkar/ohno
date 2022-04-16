package src

type Input struct {
	Url	string `json:"url"`
}

type Output struct {
	ScanId string `json:"scan_id"`
}

type ErrorType struct {
	Error bool `json:"error"`
	Message string `json:"message"`
}

type Result struct {
	ScanId string `json:"scan_id"`
	Code string `json:"code"`
	Filename string `json:"filename"`
	IssueSeverity string `json:"issue_severity"`
	CreatedAt string `json:"created_at"`
}

type Scan struct {
	Safety bool `json:"safety"`
	Results []Result `json:"results"`
}