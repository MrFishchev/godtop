package domain

type Volume struct {
	Name        string `json:"name"`
	Source      string `json:"source"`
	Destination string `json:"destination"`
	Size        int64  `json:"size"`
}
