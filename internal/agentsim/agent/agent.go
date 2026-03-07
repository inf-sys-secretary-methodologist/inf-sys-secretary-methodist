package agent

// Agent represents a virtual user that interacts with the system.
type Agent struct {
	Name        string // Full name, e.g. "Марина Петровна Соколова"
	Email       string // e.g. "m.sokolova@uni.local"
	Password    string // Generated password
	Role        string // system_admin, methodist, academic_secretary, teacher, student
	Personality string // Brief description for LLM context

	// Runtime state (populated after login)
	AccessToken  string
	RefreshToken string
	UserID       int64
}

// ShortName returns a display-friendly short name for logs.
func (a *Agent) ShortName() string {
	if a.Name != "" {
		return a.Name
	}
	return a.Email
}

// IsAuthenticated returns true if the agent has a valid access token.
func (a *Agent) IsAuthenticated() bool {
	return a.AccessToken != ""
}
