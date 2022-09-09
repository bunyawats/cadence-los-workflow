package common

type (
	DEResult struct {
		AppID  string `json:"appID"`
		Status string `json:"status"`
	}
)

type TaskToken struct {
	DomainID   string `json:"domainId"`
	WorkflowID string `json:"workflowId"`
	RunID      string `json:"runId"`
	ScheduleID int64  `json:"scheduleId"`
	State      string `json:"state"`
}

type LoanApplication struct {
	AppID     string `bson:"appID" json:"appID"`
	Fname     string `bson:"fname" json:"fname"`
	Lname     string `bson:"lname" json:"lname"`
	Email     string `bson:"email" json:"email"`
	PhoneNo   string `bson:"phoneNo" json:"phoneNo"`
	TaskToken string `bson:"taskToken" json:"taskToken"`
	LastState string `bson:"lastState" json:"lastState"`
}
