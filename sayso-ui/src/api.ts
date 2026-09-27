import type {
  ActionItem,
  AppSettings,
  AppState,
  Project,
  RecordingSegment,
  Session,
} from "./types";

import { call, upload } from "./transport";

async function request<T>(url: string, init?: RequestInit): Promise<T> {
  return call<T>(
    init?.method || "GET",
    url,
    init?.body ? JSON.parse(String(init.body)) : undefined,
  );
}

export const api = {
  cancelRun: (id: string) =>
    request(`/api/runs/${id}/cancel`, { method: "POST" }),
  state: () => request<AppState>("/api/state"),
  updateSettings: (body: Partial<AppSettings>) =>
    request<AppSettings>("/api/settings", {
      method: "PATCH",
      body: JSON.stringify(body),
    }),
  createProject: (body: Pick<Project, "name" | "rootPath">) =>
    request<Project>("/api/projects", {
      method: "POST",
      body: JSON.stringify(body),
    }),
  updateProject: (
    id: string,
    body: Partial<Pick<Project, "name" | "rootPath">>,
  ) =>
    request<Project>(`/api/projects/${id}`, {
      method: "PATCH",
      body: JSON.stringify(body),
    }),
  createSession: (projectId: string, title: string) =>
    request<Session>("/api/sessions", {
      method: "POST",
      body: JSON.stringify({ projectId, title }),
    }),
  updateSession: (
    id: string,
    body: Partial<Pick<Session, "title" | "status">>,
  ) =>
    request<Session>(`/api/sessions/${id}`, {
      method: "PATCH",
      body: JSON.stringify(body),
    }),
  deleteSession: (id: string) =>
    request<void>(`/api/sessions/${id}`, { method: "DELETE" }),
  addUtterance: (sessionId: string, body: { text: string; atMs: number }) =>
    request(`/api/sessions/${sessionId}/utterances`, {
      method: "POST",
      body: JSON.stringify(body),
    }),
  extractActions: (sessionId: string) =>
    request(`/api/sessions/${sessionId}/extract-actions`, { method: "POST" }),
  createAction: (
    sessionId: string,
    body: Pick<ActionItem, "title" | "detail" | "priority">,
  ) =>
    request<ActionItem>(`/api/sessions/${sessionId}/actions`, {
      method: "POST",
      body: JSON.stringify(body),
    }),
  updateAction: (
    id: string,
    body: Partial<Pick<ActionItem, "title" | "detail" | "priority" | "status">>,
  ) =>
    request<ActionItem>(`/api/actions/${id}`, {
      method: "PATCH",
      body: JSON.stringify(body),
    }),
  deleteAction: (id: string) =>
    request<void>(`/api/actions/${id}`, { method: "DELETE" }),
  dispatchAction: (id: string) =>
    request(`/api/actions/${id}/dispatch`, { method: "POST" }),
  uploadAudio: upload,
};
