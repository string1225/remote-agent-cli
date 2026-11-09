package protocol

// RunActivity contains bounded display metadata, never raw tool results or input bodies.
type RunActivity struct {
	ID          string     `json:"id"`
	Kind        string     `json:"kind"`
	Name        string     `json:"name"`
	Status      string     `json:"status"`
	StartedAt   int64      `json:"startedAt,omitempty"`
	CompletedAt int64      `json:"completedAt,omitempty"`
	Steps       []PlanStep `json:"steps,omitempty"`
}

type PlanStep struct {
	Text      string `json:"text"`
	Completed bool   `json:"completed"`
}

type WorkspaceChange struct {
	Path  string `json:"path"`
	Index string `json:"index"`
	Work  string `json:"work"`
}

type WorkspaceChanges struct {
	Git       bool              `json:"git"`
	Branch    string            `json:"branch,omitempty"`
	Files     []WorkspaceChange `json:"files"`
	Truncated bool              `json:"truncated,omitempty"`
}

type WorkspaceDiff struct {
	Path      string `json:"path"`
	Staged    string `json:"staged,omitempty"`
	Unstaged  string `json:"unstaged,omitempty"`
	Untracked string `json:"untracked,omitempty"`
	Binary    bool   `json:"binary,omitempty"`
	Truncated bool   `json:"truncated,omitempty"`
}
