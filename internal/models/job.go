package models

const (
	StatusPending  = "pending"
	StatusApproved = "approved"
	StatusRejected = "rejected"
)

type JobApp struct {
	ID          int64  `json:"id"`
	ClientID    int64  `json:"client_id"`
	JobTitle    string `json:"job_title"`
	Description string `json:"description"`
	Status      string `json:"status"`
	Category    string `json:"category"`
	Location    string `json:"location"`
}

type CreateJobRequest struct {
	JobTitle    string `json:"job_title"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Location    string `json:"location"`
}

type UpdateJobRequest struct {
	JobTitle    string `json:"job_title,omitempty"`
	Description string `json:"description,omitempty"`
	Category    string `json:"category,omitempty"`
	Location    string `json:"location,omitempty"`
}

type UpdateJobStatusRequest struct {
	Status string `json:"status"`
}
