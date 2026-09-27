package agent

// SaySo's local data model is kept compatible with its original discussion UI.
type soProject struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	RootPath  string `json:"rootPath"`
	Color     string `json:"color"`
	CreatedAt string `json:"createdAt"`
}
type soUtterance struct {
	ID        string `json:"id"`
	Text      string `json:"text"`
	AtMS      int64  `json:"atMs"`
	CreatedAt string `json:"createdAt"`
}
type soRecording struct {
	ID         string `json:"id"`
	Filename   string `json:"filename"`
	StartedAt  string `json:"startedAt"`
	EndedAt    string `json:"endedAt"`
	DurationMS int64  `json:"durationMs"`
	SizeBytes  int64  `json:"sizeBytes"`
	MimeType   string `json:"mimeType"`
}
type soTrace struct {
	ID               string   `json:"id"`
	Status           string   `json:"status"`
	StartedAt        string   `json:"startedAt"`
	FinishedAt       string   `json:"finishedAt,omitempty"`
	Model            string   `json:"model"`
	ContextLabel     string   `json:"contextLabel"`
	Input            string   `json:"input"`
	Output           string   `json:"output"`
	Events           []string `json:"events"`
	UtteranceCount   int      `json:"utteranceCount"`
	ActionCount      int      `json:"actionCount"`
	CreatedActionIDs []string `json:"createdActionIds"`
	UpdatedActionIDs []string `json:"updatedActionIds"`
	Error            string   `json:"error,omitempty"`
}
type soSession struct {
	ID                      string        `json:"id"`
	ProjectID               string        `json:"projectId"`
	Title                   string        `json:"title"`
	Status                  string        `json:"status"`
	StartedAt               string        `json:"startedAt"`
	Recordings              []soRecording `json:"recordings"`
	Transcript              []soUtterance `json:"transcript"`
	AnalysisStatus          string        `json:"analysisStatus"`
	AnalysisError           string        `json:"analysisError,omitempty"`
	AnalysisCursor          int           `json:"analysisCursor"`
	AnalysisCompletedSlices int           `json:"analysisCompletedSlices"`
	AnalysisNextAt          string        `json:"analysisNextAt,omitempty"`
	LastAnalyzedAt          string        `json:"lastAnalyzedAt,omitempty"`
	AnalysisActiveTraceID   string        `json:"analysisActiveTraceId,omitempty"`
	AnalysisTraces          []*soTrace    `json:"analysisTraces"`
	owner                   *peer
	timerCancel             func()
	pending                 bool
	pendingFull             bool
}
type soAction struct {
	ID        string `json:"id"`
	SessionID string `json:"sessionId"`
	ProjectID string `json:"projectId"`
	Title     string `json:"title"`
	Detail    string `json:"detail"`
	Source    string `json:"source"`
	Priority  string `json:"priority"`
	Status    string `json:"status"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
	RunID     string `json:"runId,omitempty"`
}
type soRun struct {
	ID         string `json:"id"`
	ActionID   string `json:"actionId"`
	Status     string `json:"status"`
	StartedAt  string `json:"startedAt"`
	FinishedAt string `json:"finishedAt,omitempty"`
	Output     string `json:"output"`
	Error      string `json:"error,omitempty"`
	cancel     func()
}
type soSettings struct {
	AnalysisProvider        string `json:"analysisProvider"`
	AnalysisModel           string `json:"analysisModel"`
	AnalysisIntervalMinutes int    `json:"analysisIntervalMinutes"`
	Executor                string `json:"executor"`
	ExecutionModel          string `json:"executionModel"`
}
type soState struct {
	Projects []*soProject `json:"projects"`
	Sessions []*soSession `json:"sessions"`
	Actions  []*soAction  `json:"actions"`
	Runs     []*soRun     `json:"runs"`
	Settings soSettings   `json:"settings"`
}
