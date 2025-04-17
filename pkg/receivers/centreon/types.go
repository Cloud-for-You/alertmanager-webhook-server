package centreon

// CheckResult represents the structure of a single check result in JSON format.
type CheckResult struct {
	CheckResult struct {
		Type string `json:"type"` // Must be set, e.g., "service"
	} `json:"checkresult"`
	Hostname    string `json:"hostname"`    // Must be set
	Servicename string `json:"servicename,omitempty"` // Optional
	State       string `json:"state"`       // Must be set, e.g., "0", "1", "2"
	Output      string `json:"output"`      // Must be set
}

// CheckResultsPayload represents the payload containing multiple check results.
type CheckResultsPayload struct {
	CheckResults []CheckResult `json:"checkresults"`
}