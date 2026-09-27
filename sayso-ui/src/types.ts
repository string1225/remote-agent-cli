export type SessionStatus = "paused" | "recording";
export type ActionStatus = "draft" | "ready" | "running" | "done" | "failed";
export type ActionPriority = "high" | "medium" | "low";

export interface Project {
  id: string;
  name: string;
  rootPath: string;
  color: string;
  createdAt: string;
}

export interface Utterance {
  id: string;
  speaker?: string;
  text: string;
  atMs: number;
  createdAt: string;
}

export interface RecordingSegment {
  id: string;
  filename: string;
  startedAt: string;
  endedAt: string;
  durationMs?: number;
  sizeBytes: number;
  mimeType: string;
}

export interface AnalysisTrace {
  id: string;
  status: "queued" | "running" | "done" | "failed";
  startedAt: string;
  finishedAt?: string;
  model: string;
  contextLabel: string;
  input: string;
  output: string;
  events: string[];
  utteranceCount: number;
  actionCount: number;
  createdActionIds: string[];
  updatedActionIds: string[];
  error?: string;
}

export interface Session {
  id: string;
  projectId: string;
  title: string;
  status: SessionStatus;
  startedAt: string;
  recordingPath?: string;
  recordingPaths?: string[];
  recordings: RecordingSegment[];
  transcript: Utterance[];
  analysisStatus?: "idle" | "queued" | "running" | "done" | "failed";
  analysisError?: string;
  analysisPending?: boolean;
  analysisPendingFull?: boolean;
  analysisCursor?: number;
  analysisCompletedSlices?: number;
  analysisNextAt?: string;
  lastAnalyzedAt?: string;
  analysisActiveTraceId?: string;
  analysisTraces?: AnalysisTrace[];
}

export interface AppSettings {
  analysisProvider: "codex" | "local";
  analysisModel: string;
  analysisIntervalMinutes: 2;
  executor: "codex";
  executionModel: string;
}

export interface ActionItem {
  id: string;
  sessionId: string;
  projectId: string;
  title: string;
  detail: string;
  source: string;
  priority: ActionPriority;
  status: ActionStatus;
  createdAt: string;
  updatedAt: string;
  runId?: string;
}

export interface CodexRun {
  id: string;
  actionId: string;
  status: "running" | "done" | "failed";
  startedAt: string;
  finishedAt?: string;
  output: string;
  error?: string;
}

export interface AppState {
  projects: Project[];
  sessions: Session[];
  actions: ActionItem[];
  runs: CodexRun[];
  settings: AppSettings;
}
