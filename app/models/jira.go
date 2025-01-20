package models

type JiraResponse struct {
	Id string `json:"id"`
}

type AttachmentInfo struct {
	Content  string
	MimeType string
}

type JiraTask struct {
	Key    string `json:"key"`
	Fields struct {
		Summary string `json:"summary"`
		Status  struct {
			Name string `json:"name"`
		} `json:"status"`
		Assignee *struct {
			DisplayName string `json:"displayName"`
		} `json:"assignee"`
	} `json:"fields"`
}
