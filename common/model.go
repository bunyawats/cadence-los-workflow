package common

const (
	ApplicationName            = "loanOnBoarding"
	LoanOnBoardingWorkflowName = "loanOnBoardingWorkflow"

	SignalName = "trigger-signal"
	QueryName  = "state"
)

const (
	Create                  Action = "create"
	SubmitFormOne                  = "submitFormOne"
	SubmitFormTwo                  = "submitFormTwo"
	SubmitDEOne                    = "submitDEOne"
	DEOneResultNotification        = "deOneResultNotification"
	Approve                        = "approve"
	Reject                         = "reject"
	Cancel                         = "cancel"
)

const (
	Initialized      State = "initialized"
	Received               = "received"
	Created                = "created"
	FormOneSubmitted       = "formOneSubmitted"
	FormTwoSubmitted       = "formTwoSubmitted"
	DEOneSubmitted         = "deOneSubmitted"
	Approved               = "approved"
	Rejected               = "rejected"
	Canceled               = "canceled"
)

type (
	DEResult struct {
		AppID  string `json:"appID"`
		Status string `json:"status"`
	}

	TaskToken struct {
		DomainID   string `json:"domainId"`
		WorkflowID string `json:"workflowId"`
		RunID      string `json:"runId"`
		ScheduleID int64  `json:"scheduleId"`
		State      string `json:"state"`
	}

	LoanApplication struct {
		AppID     string `bson:"appID" json:"appID"`
		Fname     string `bson:"fname" json:"fname"`
		Lname     string `bson:"lname" json:"lname"`
		Email     string `bson:"email" json:"email"`
		PhoneNo   string `bson:"phoneNo" json:"phoneNo"`
		TaskToken string `bson:"taskToken" json:"taskToken"`
		LastState string `bson:"lastState" json:"lastState"`
	}

	State  string
	Action string

	QueryResult struct {
		State   State
		Content string
	}

	SignalPayload struct {
		Action  Action
		Content string
	}
)
