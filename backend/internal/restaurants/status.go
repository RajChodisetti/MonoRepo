package restaurants

const (
	StatusLead             = "lead"
	StatusDemoReady        = "demo_ready"
	StatusEmailed          = "emailed"
	StatusInterested       = "interested"
	StatusClientOnboarding = "client_onboarding"
	StatusActiveClient     = "active_client"
	StatusLost             = "lost"
	StatusArchived         = "archived"
)

func ValidStatuses() []string {
	return []string{
		StatusLead,
		StatusDemoReady,
		StatusEmailed,
		StatusInterested,
		StatusClientOnboarding,
		StatusActiveClient,
		StatusLost,
		StatusArchived,
	}
}

func IsValidStatus(status string) bool {
	for _, candidate := range ValidStatuses() {
		if candidate == status {
			return true
		}
	}
	return false
}
