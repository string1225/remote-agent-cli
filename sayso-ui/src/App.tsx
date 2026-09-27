import {
  AudioLines,
  Check,
  CheckCircle2,
  ChevronDown,
  ChevronRight,
  ChevronUp,
  Circle,
  Clock3,
  Code2,
  FileAudio,
  FolderGit2,
  Settings2,
  LoaderCircle,
  Mic,
  MoreHorizontal,
  Pause,
  PencilLine,
  Play,
  Plus,
  Radio,
  RefreshCw,
  Search,
  Send,
  Sparkles,
  Trash2,
  X,
  Zap,
} from "lucide-react";
import { FormEvent, useEffect, useMemo, useRef, useState } from "react";
import { api } from "./api";
import { recordingURL } from "./transport";
import {
  normalizeDiscussionSearch,
  splitDiscussionSearchMatches,
  textMatchesDiscussionSearch,
} from "./discussionSearch";
import type {
  ActionItem,
  ActionPriority,
  AnalysisTrace,
  AppSettings,
  AppState,
  Project,
  RecordingSegment,
  Session,
} from "./types";

const MAX_RECORDING_SEGMENT_MS = 10 * 60 * 1000;
const defaultSettings: AppSettings = {
  analysisProvider: "codex",
  analysisModel: "default",
  analysisIntervalMinutes: 2,
  executor: "codex",
  executionModel: "default",
};
const emptyState: AppState = {
  projects: [],
  sessions: [],
  actions: [],
  runs: [],
  settings: defaultSettings,
};
const modelOptions = [
  { value: "gpt-5.4-mini", label: "GPT-5.4 mini（需账号支持）" },
  { value: "gpt-5.6-luna", label: "GPT-5.6 Luna" },
  { value: "gpt-5.6-terra", label: "GPT-5.6 Terra" },
  { value: "gpt-5.6-sol", label: "GPT-5.6 Sol" },
  { value: "default", label: "Codex 默认模型（推荐）" },
];

const executionModelOptions = [
  { value: "default", label: "Codex 默认模型（推荐）" },
  { value: "gpt-5.6-sol", label: "GPT-5.6 Sol" },
  { value: "gpt-5.6-terra", label: "GPT-5.6 Terra" },
  { value: "gpt-5.6-luna", label: "GPT-5.6 Luna" },
  { value: "gpt-5.4-mini", label: "GPT-5.4 mini" },
];

function formatClock(milliseconds: number) {
  const seconds = Math.max(0, Math.floor(milliseconds / 1000));
  return `${String(Math.floor(seconds / 60)).padStart(2, "0")}:${String(seconds % 60).padStart(2, "0")}`;
}

function systemMinuteKey(iso: string) {
  const date = new Date(iso);
  if (Number.isNaN(date.getTime())) return iso;
  return `${date.getFullYear()}-${date.getMonth()}-${date.getDate()}-${date.getHours()}-${date.getMinutes()}`;
}

function formatSystemMinute(iso: string) {
  const date = new Date(iso);
  if (Number.isNaN(date.getTime())) return "--:--";
  return `${String(date.getHours()).padStart(2, "0")}:${String(date.getMinutes()).padStart(2, "0")}`;
}

function formatBytes(bytes: number) {
  if (!bytes) return "大小未知";
  if (bytes < 1024 * 1024) return `${Math.max(1, Math.round(bytes / 1024))} KB`;
  return `${(bytes / 1024 / 1024).toFixed(1)} MB`;
}

function relativeDay(iso: string) {
  const date = new Date(iso);
  const today = new Date();
  if (date.toDateString() === today.toDateString()) return "今天";
  const yesterday = new Date(today);
  yesterday.setDate(today.getDate() - 1);
  if (date.toDateString() === yesterday.toDateString()) return "昨天";
  return `${date.getMonth() + 1}月${date.getDate()}日`;
}

function Modal({
  title,
  description,
  children,
  onClose,
  wide = false,
}: {
  title: string;
  description?: string;
  children: React.ReactNode;
  onClose: () => void;
  wide?: boolean;
}) {
  return (
    <div
      className="modal-backdrop"
      onMouseDown={(event) => event.target === event.currentTarget && onClose()}
    >
      <section
        className={`modal ${wide ? "modal-wide" : ""}`}
        role="dialog"
        aria-modal="true"
      >
        <div className="modal-heading">
          <div>
            <h2>{title}</h2>
            {description && <p>{description}</p>}
          </div>
          <button className="icon-button" onClick={onClose} aria-label="关闭">
            <X size={18} />
          </button>
        </div>
        {children}
      </section>
    </div>
  );
}

function StatusPill({ status }: { status: ActionItem["status"] }) {
  const copy = {
    draft: "待完善",
    ready: "Ready",
    running: "Codex 工作中",
    done: "已完成",
    failed: "执行失败",
  }[status];
  return (
    <span className={`status-pill status-${status}`}>
      {status === "running" && <LoaderCircle size={12} />} {copy}
    </span>
  );
}

function useTypewriterLines(lines: string[], active: boolean) {
  const [typedLines, setTypedLines] = useState<string[]>([]);
  const lineSignature = lines.join("\u0000");

  useEffect(() => {
    if (!active) {
      setTypedLines([]);
      return undefined;
    }

    if (window.matchMedia("(prefers-reduced-motion: reduce)").matches) {
      setTypedLines(lines);
      return undefined;
    }

    let cancelled = false;
    let lineIndex = 0;
    let characterIndex = 0;
    let timer: number | undefined;
    setTypedLines([]);

    const typeNextCharacter = () => {
      if (cancelled) return;
      const characters = Array.from(lines[lineIndex]);
      if (characterIndex < characters.length) {
        characterIndex += 1;
        const nextText = characters.slice(0, characterIndex).join("");
        setTypedLines((current) => {
          const next = current.slice();
          next[lineIndex] = nextText;
          return next;
        });
        timer = window.setTimeout(typeNextCharacter, 34);
        return;
      }
      if (lineIndex < lines.length - 1) {
        lineIndex += 1;
        characterIndex = 0;
        timer = window.setTimeout(typeNextCharacter, 360);
        return;
      }
      timer = window.setTimeout(() => {
        lineIndex = 0;
        characterIndex = 0;
        setTypedLines([]);
        timer = window.setTimeout(typeNextCharacter, 240);
      }, 1600);
    };

    timer = window.setTimeout(typeNextCharacter, 160);
    return () => {
      cancelled = true;
      if (timer !== undefined) window.clearTimeout(timer);
    };
  }, [active, lineSignature]);

  return typedLines;
}

export default function App({
  onClose,
  deviceLabel,
}: {
  onClose: () => void;
  deviceLabel: string;
}) {
  const [closing, setClosing] = useState(false);
  const [pendingAudio, setPendingAudio] = useState<
    Array<{ url: string; name: string }>
  >([]);
  const recoveryRef = useRef<Array<{ url: string; name: string }>>([]);
  const uploadsRef = useRef(new Set<Promise<unknown>>());
  const stopRef = useRef<() => Promise<void>>(async () => {});
  const [manualText, setManualText] = useState("");
  const [manualBusy, setManualBusy] = useState(false);
  const [mobilePane, setMobilePane] = useState<"discussion" | "actions">(
    "discussion",
  );
  const [mobileTree, setMobileTree] = useState(false);
  useEffect(() => {
    const beforeUnload = (event: BeforeUnloadEvent) => {
      if (recorderRef.current || uploadsRef.current.size) {
        event.preventDefault();
        event.returnValue = "";
      }
    };
    const disconnected = () => {
      void stopRef
        .current()
        .catch(() => {})
        .finally(() => showToast("设备连接已断开，未上传的录音可下载保存。"));
    };
    window.addEventListener("beforeunload", beforeUnload);
    window.addEventListener("remote-agent-disconnected", disconnected);
    return () => {
      recognitionFatalRef.current = true;
      try {
        recognitionRef.current?.stop();
      } catch {}
      streamRef.current?.getTracks().forEach((track) => track.stop());
      if (rotationTimerRef.current)
        window.clearTimeout(rotationTimerRef.current);
      recoveryRef.current.forEach((item) => URL.revokeObjectURL(item.url));
      window.removeEventListener("beforeunload", beforeUnload);
      window.removeEventListener("remote-agent-disconnected", disconnected);
    };
  }, []);
  async function closeWorkspace() {
    setClosing(true);
    try {
      await stopRef.current();
      await Promise.allSettled([...uploadsRef.current]);
      if (
        recoveryRef.current.length &&
        !confirm("仍有未上传录音，请先下载保存。确定离开吗？")
      )
        return;
      onClose();
    } catch (error) {
      showToast(
        error instanceof Error ? error.message : "正在保存，请稍后再试",
      );
    } finally {
      setClosing(false);
    }
  }

  const [state, setState] = useState<AppState>(emptyState);
  const [loading, setLoading] = useState(true);
  const [selectedProjectId, setSelectedProjectId] = useState("");
  const [selectedSessionId, setSelectedSessionId] = useState("");
  const [collapsedProjects, setCollapsedProjects] = useState<Set<string>>(
    new Set(),
  );
  const [projectModal, setProjectModal] = useState<Project | "new" | null>(
    null,
  );
  const [systemModal, setSystemModal] = useState(false);
  const [renameSession, setRenameSession] = useState<Session | null>(null);
  const [actionModal, setActionModal] = useState<ActionItem | "new" | null>(
    null,
  );
  const [runModal, setRunModal] = useState<ActionItem | null>(null);
  const [recordingsModal, setRecordingsModal] = useState<Session | null>(null);
  const [toast, setToast] = useState("");
  const [search, setSearch] = useState("");
  const [interimText, setInterimText] = useState("");
  const [isRecording, setIsRecording] = useState(false);
  const [recordingSessionId, setRecordingSessionId] = useState("");
  const [elapsed, setElapsed] = useState(0);
  const recorderRef = useRef<MediaRecorder | null>(null);
  const recognitionRef = useRef<{ start: () => void; stop: () => void } | null>(
    null,
  );
  const recognitionFatalRef = useRef(false);
  const streamRef = useRef<MediaStream | null>(null);
  const rotationTimerRef = useRef<number | null>(null);
  const activeSessionRef = useRef<Session | null>(null);
  const pollRef = useRef<number | null>(null);

  function showToast(message: string) {
    setToast(message);
    window.setTimeout(() => setToast(""), 3600);
  }

  const refresh = async (quiet = false) => {
    try {
      const next = await api.state();
      setState(next);
      setSelectedProjectId((current) =>
        current && next.projects.some((item) => item.id === current)
          ? current
          : next.projects[0]?.id || "",
      );
      setSelectedSessionId((current) =>
        current && next.sessions.some((item) => item.id === current)
          ? current
          : next.sessions.find(
              (item) =>
                item.projectId === (selectedProjectId || next.projects[0]?.id),
            )?.id || "",
      );
    } catch (error) {
      if (!quiet)
        showToast(error instanceof Error ? error.message : "无法连接服务");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void refresh();
  }, []);

  const hasBackgroundWork =
    state.actions.some((action) => action.status === "running") ||
    state.sessions.some(
      (session) =>
        session.status === "recording" ||
        ["queued", "running"].includes(session.analysisStatus || ""),
    );
  useEffect(() => {
    if (hasBackgroundWork && !pollRef.current)
      pollRef.current = window.setInterval(() => void refresh(true), 1800);
    if (!hasBackgroundWork && pollRef.current) {
      window.clearInterval(pollRef.current);
      pollRef.current = null;
    }
    return () => {
      if (pollRef.current) window.clearInterval(pollRef.current);
      pollRef.current = null;
    };
  }, [hasBackgroundWork]);

  const selectedProject =
    state.projects.find((project) => project.id === selectedProjectId) ||
    state.projects[0];
  const selectedSession = state.sessions.find(
    (session) => session.id === selectedSessionId,
  );
  const sessionActions = state.actions
    .filter((action) => action.sessionId === selectedSession?.id)
    .sort((a, b) => {
      const rank = { running: 0, ready: 1, draft: 2, failed: 3, done: 4 };
      return (
        rank[a.status] - rank[b.status] ||
        b.createdAt.localeCompare(a.createdAt)
      );
    });
  const sessionsByProject = useMemo(
    () =>
      new Map(
        state.projects.map((project) => [
          project.id,
          state.sessions.filter((session) => session.projectId === project.id),
        ]),
      ),
    [state.projects, state.sessions],
  );

  useEffect(() => {
    if (!isRecording || !activeSessionRef.current) return;
    const update = () =>
      setElapsed(
        Date.now() - new Date(activeSessionRef.current!.startedAt).getTime(),
      );
    update();
    const timer = window.setInterval(update, 1000);
    return () => window.clearInterval(timer);
  }, [isRecording]);

  function selectProject(id: string) {
    setMobileTree(false);
    setSelectedProjectId(id);
    setSelectedSessionId(
      state.sessions.find((session) => session.projectId === id)?.id || "",
    );
    setCollapsedProjects((current) => {
      const next = new Set(current);
      next.delete(id);
      return next;
    });
  }

  async function createSession() {
    if (!selectedProject) return;
    try {
      if (isRecording) await pauseCapture();
      const session = await api.createSession(selectedProject.id, "");
      setSelectedSessionId(session.id);
      setMobileTree(false);
      await refresh(true);
    } catch (error) {
      showToast(error instanceof Error ? error.message : "创建会话失败");
    }
  }

  function startRecorderSegment(stream: MediaStream, session: Session) {
    const startedAtMs = Date.now();
    const chunks: Blob[] = [];
    const recorder = new MediaRecorder(stream, { audioBitsPerSecond: 64000 });
    recorderRef.current = recorder;
    recorder.ondataavailable = (event) => {
      if (event.data.size) chunks.push(event.data);
    };
    let segmentResolve: () => void = () => {};
    const segmentDone = new Promise<void>((resolve) => {
      segmentResolve = resolve;
    });
    (recorder as MediaRecorder & { saved?: Promise<void> }).saved = segmentDone;
    recorder.onstop = () => {
      const endedAtMs = Date.now();
      const blob = new Blob(chunks, {
        type: recorder.mimeType || "audio/webm",
      });
      if (!blob.size) {
        segmentResolve();
        return;
      }
      const task = api
        .uploadAudio(session.id, blob, {
          startedAt: new Date(startedAtMs).toISOString(),
          endedAt: new Date(endedAtMs).toISOString(),
          durationMs: endedAtMs - startedAtMs,
        })
        .then(() => refresh(true))
        .catch((error) => {
          const url = URL.createObjectURL(blob);
          recoveryRef.current.push({
            url,
            name: `sayso-${new Date(startedAtMs).toISOString().replace(/:/g, "-")}.${blob.type.includes("mp4") ? "m4a" : "webm"}`,
          });
          setPendingAudio([...recoveryRef.current]);
          showToast(`录音未上传：${error.message}。请下载备份。`);
        })
        .finally(() => {
          uploadsRef.current.delete(task);
          segmentResolve();
        });
      uploadsRef.current.add(task);
    };
    recorder.start(1000);
    rotationTimerRef.current = window.setTimeout(
      () => rotateRecordingSegment(session),
      MAX_RECORDING_SEGMENT_MS,
    );
  }

  function rotateRecordingSegment(session: Session) {
    if (rotationTimerRef.current) window.clearTimeout(rotationTimerRef.current);
    rotationTimerRef.current = null;
    const stream = streamRef.current;
    const recorder = recorderRef.current;
    if (!stream || !recorder || recorder.state !== "recording") return;
    recorder.stop();
    startRecorderSegment(stream, session);
  }

  async function beginCapture(session: Session) {
    if (isRecording) await pauseCapture();
    if (!navigator.mediaDevices?.getUserMedia) {
      showToast("当前浏览器不支持录音");
      return;
    }
    try {
      const stream = await navigator.mediaDevices.getUserMedia({
        audio: { echoCancellation: true, noiseSuppression: true },
      });
      streamRef.current = stream;
      await api.updateSession(session.id, { status: "recording" });
      activeSessionRef.current = session;
      recognitionFatalRef.current = false;
      startRecorderSegment(stream, session);

      const speechWindow = window as unknown as {
        SpeechRecognition?: new () => any;
        webkitSpeechRecognition?: new () => any;
      };
      const Recognition =
        speechWindow.SpeechRecognition || speechWindow.webkitSpeechRecognition;
      if (Recognition) {
        const recognition = new Recognition();
        recognition.continuous = true;
        recognition.interimResults = true;
        recognition.lang = "zh-CN";
        recognition.onresult = (event: any) => {
          let interim = "";
          for (
            let index = event.resultIndex;
            index < event.results.length;
            index += 1
          ) {
            const result = event.results[index];
            const text = result[0].transcript.trim();
            if (result.isFinal && text) {
              void api
                .addUtterance(session.id, {
                  text,
                  atMs: Date.now() - new Date(session.startedAt).getTime(),
                })
                .then(() => refresh(true))
                .catch((error) => showToast(error.message));
            } else interim += text;
          }
          setInterimText(interim);
        };
        recognition.onerror = (event: { error: string }) => {
          if (
            [
              "network",
              "service-not-allowed",
              "not-allowed",
              "audio-capture",
            ].includes(event.error)
          ) {
            recognitionFatalRef.current = true;
            setInterimText("");
            showToast(
              event.error === "network"
                ? "实时转写服务网络不可用；录音仍会继续并正常保存"
                : "录音正常；当前浏览器暂时无法使用实时转写",
            );
          }
        };
        recognition.onend = () => {
          if (
            !recognitionFatalRef.current &&
            recorderRef.current?.state === "recording"
          ) {
            try {
              recognition.start();
            } catch {
              /* browser is still restarting */
            }
          }
        };
        try {
          recognition.start();
        } catch {
          recognitionFatalRef.current = true;
          showToast("录音已开始，实时转写不可用，可以打字补充。");
        }
        recognitionRef.current = recognition;
      } else showToast("录音已开始；当前浏览器不提供实时转写");

      setRecordingSessionId(session.id);
      setIsRecording(true);
      setElapsed(Date.now() - new Date(session.startedAt).getTime());
      await refresh(true);
    } catch (error) {
      if (rotationTimerRef.current)
        window.clearTimeout(rotationTimerRef.current);
      if (recorderRef.current?.state === "recording")
        recorderRef.current.stop();
      await api.updateSession(session.id, { status: "paused" }).catch(() => {});
      streamRef.current?.getTracks().forEach((track) => track.stop());
      streamRef.current = null;
      recorderRef.current = null;
      showToast(
        error instanceof Error
          ? `无法开始录音：${error.message}`
          : "无法开始录音",
      );
    }
  }

  async function pauseCapture() {
    const sessionId = activeSessionRef.current?.id || recordingSessionId;
    const saved = (
      recorderRef.current as (MediaRecorder & { saved?: Promise<void> }) | null
    )?.saved;
    recognitionFatalRef.current = true;
    try {
      recognitionRef.current?.stop();
    } catch {
      /* already stopped */
    }
    recognitionRef.current = null;
    if (rotationTimerRef.current) window.clearTimeout(rotationTimerRef.current);
    rotationTimerRef.current = null;
    if (recorderRef.current?.state !== "inactive") recorderRef.current?.stop();
    streamRef.current?.getTracks().forEach((track) => track.stop());
    streamRef.current = null;
    recorderRef.current = null;
    activeSessionRef.current = null;
    setInterimText("");
    setIsRecording(false);
    setRecordingSessionId("");
    await saved;
    await Promise.allSettled([...uploadsRef.current]);
    if (sessionId) {
      await api.updateSession(sessionId, { status: "paused" });
      await refresh(true);
    }
  }

  stopRef.current = pauseCapture;

  async function mutateAction(
    callback: () => Promise<unknown>,
    success?: string,
  ) {
    try {
      await callback();
      if (success) showToast(success);
      await refresh(true);
    } catch (error) {
      showToast(error instanceof Error ? error.message : "操作失败");
    }
  }

  if (loading)
    return (
      <div className="loading-screen">
        <div className="brand-mark">
          <AudioLines size={22} />
        </div>
        <p>正在打开 SaySo…</p>
        <button onClick={onClose}>返回工作台</button>
      </div>
    );

  return (
    <div
      className={`sayso-layout pane-${mobilePane} ${mobileTree ? "tree-open" : ""}`}
    >
      <header className="remote-header">
        <button onClick={() => void closeWorkspace()} disabled={closing}>
          ← {closing ? "正在保存…" : "工作台"}
        </button>
        <strong>SaySo</strong>
        <span>{deviceLabel}</span>
        <button
          className="mobile-tree-button"
          onClick={() => setMobileTree(!mobileTree)}
        >
          项目 / 会话
        </button>
      </header>
      {pendingAudio.length > 0 && (
        <div className="audio-recovery">
          未上传录音，请下载保存：
          {pendingAudio.map((item) => (
            <a key={item.url} href={item.url} download={item.name}>
              下载录音备份
            </a>
          ))}
        </div>
      )}
      <div className="app-shell">
        <aside className="sidebar">
          <div className="brand-row">
            <div className="brand">
              <div className="brand-mark">
                <AudioLines size={21} />
              </div>
              <span>SaySo</span>
            </div>
            <button
              className="system-settings"
              onClick={() => setSystemModal(true)}
              aria-label="系统设置"
            >
              <Settings2 size={17} />
            </button>
          </div>
          <button
            className="new-session-button"
            onClick={() => void createSession()}
            disabled={!selectedProject}
          >
            <Plus size={16} /> 新会话 <span>⌘ N</span>
          </button>
          <div className="tree-heading">
            <span>项目</span>
            <button
              onClick={() => setProjectModal("new")}
              aria-label="添加项目"
            >
              <Plus size={14} />
            </button>
          </div>
          <div className="project-tree">
            {state.projects.map((project) => {
              const sessions = sessionsByProject.get(project.id) || [];
              const collapsed = collapsedProjects.has(project.id);
              return (
                <div className="project-branch" key={project.id}>
                  <div
                    className={`project-node ${selectedProject?.id === project.id ? "selected" : ""}`}
                  >
                    <button
                      className="project-select"
                      onClick={() => selectProject(project.id)}
                    >
                      <span
                        className="tree-chevron"
                        onClick={(event) => {
                          event.stopPropagation();
                          setCollapsedProjects((current) => {
                            const next = new Set(current);
                            if (next.has(project.id)) next.delete(project.id);
                            else next.add(project.id);
                            return next;
                          });
                        }}
                      >
                        {collapsed ? (
                          <ChevronRight size={13} />
                        ) : (
                          <ChevronDown size={13} />
                        )}
                      </span>
                      <FolderGit2 size={15} />
                      <strong>{project.name}</strong>
                    </button>
                    <button
                      className="tree-edit"
                      onClick={() => setProjectModal(project)}
                      aria-label={`编辑项目 ${project.name}`}
                    >
                      <PencilLine size={13} />
                    </button>
                  </div>
                  {!collapsed && (
                    <div className="session-tree">
                      {sessions.map((session) => {
                        const activeRecording =
                          isRecording && recordingSessionId === session.id;
                        return (
                          <div
                            className={`tree-session ${selectedSession?.id === session.id ? "active" : ""}`}
                            key={session.id}
                          >
                            <button
                              className="session-select"
                              onClick={() => {
                                setSelectedProjectId(project.id);
                                setSelectedSessionId(session.id);
                                setMobileTree(false);
                              }}
                            >
                              <span
                                className={`tree-session-icon ${activeRecording ? "recording" : ""}`}
                              >
                                {activeRecording ? (
                                  <Radio size={12} />
                                ) : (
                                  <FileAudio size={12} />
                                )}
                              </span>
                              <span>
                                <strong>{session.title}</strong>
                                <small>
                                  {relativeDay(session.startedAt)} ·{" "}
                                  {session.transcript.length} 条
                                </small>
                              </span>
                            </button>
                            <button
                              className="tree-edit"
                              onClick={() => setRenameSession(session)}
                              aria-label={`重命名 ${session.title}`}
                            >
                              <PencilLine size={12} />
                            </button>
                          </div>
                        );
                      })}
                      {sessions.length === 0 && (
                        <div className="tree-empty">暂无会话</div>
                      )}
                    </div>
                  )}
                </div>
              );
            })}
          </div>
        </aside>

        <main className="workspace">
          {!selectedSession ? (
            <EmptyWorkspace
              project={selectedProject}
              onStart={() => void createSession()}
              onProject={() => setProjectModal("new")}
            />
          ) : (
            <>
              <header className="topbar">
                <div className="title-stack">
                  <div className="eyebrow">
                    <span>{selectedProject?.name}</span>
                    <span>/</span>
                    <span>{relativeDay(selectedSession.startedAt)}</span>
                  </div>
                  <div className="session-title-line">
                    <h1>{selectedSession.title}</h1>
                    <button
                      className="small-icon"
                      onClick={() => setRenameSession(selectedSession)}
                      aria-label="编辑会话名称"
                    >
                      <PencilLine size={14} />
                    </button>
                  </div>
                </div>
                <div className="topbar-actions">
                  {(selectedSession.recordings?.length ||
                    selectedSession.recordingPath) && (
                    <button
                      className="ghost-button"
                      onClick={() => setRecordingsModal(selectedSession)}
                    >
                      <Play size={15} /> 录音片段
                    </button>
                  )}
                  {isRecording && recordingSessionId === selectedSession.id ? (
                    <div className="recording-control">
                      <span className="recording-time">
                        <i /> {formatClock(elapsed)}
                      </span>
                      <button
                        className="pause-button"
                        onClick={() =>
                          void pauseCapture().catch((error) =>
                            showToast(error.message),
                          )
                        }
                      >
                        <Pause size={13} fill="currentColor" /> 暂停
                      </button>
                    </div>
                  ) : (
                    <button
                      className="record-button"
                      onClick={() => void beginCapture(selectedSession)}
                    >
                      <Mic size={15} /> 继续录音
                    </button>
                  )}
                  <button className="icon-button" aria-label="更多">
                    <MoreHorizontal size={19} />
                  </button>
                </div>
              </header>
              <div className="mobile-tabs">
                <button
                  className={mobilePane === "discussion" ? "active" : ""}
                  onClick={() => setMobilePane("discussion")}
                >
                  讨论记录
                </button>
                <button
                  className={mobilePane === "actions" ? "active" : ""}
                  onClick={() => setMobilePane("actions")}
                >
                  Action · {sessionActions.length}
                </button>
              </div>
              <div className="content-grid">
                <TranscriptPanel
                  session={selectedSession}
                  interimText={
                    recordingSessionId === selectedSession.id ? interimText : ""
                  }
                  isRecording={
                    isRecording && recordingSessionId === selectedSession.id
                  }
                  search={search}
                  setSearch={setSearch}
                />
                <ActionPanel
                  session={selectedSession}
                  actions={sessionActions}
                  runs={state.runs}
                  settings={state.settings}
                  onCreate={() => setActionModal("new")}
                  onExtract={() =>
                    mutateAction(
                      () => api.extractActions(selectedSession.id),
                      "已开始重新分析讨论",
                    )
                  }
                  onEdit={(action) => setActionModal(action)}
                  onDelete={(action) => {
                    if (window.confirm(`删除“${action.title}”？`))
                      void mutateAction(() => api.deleteAction(action.id));
                  }}
                  onReady={(action) =>
                    mutateAction(() =>
                      api.updateAction(action.id, {
                        status: action.status === "ready" ? "draft" : "ready",
                      }),
                    )
                  }
                  onDispatch={(action) =>
                    mutateAction(
                      () => api.dispatchAction(action.id),
                      "已把 Action 交给 Codex",
                    )
                  }
                  onViewRun={(action) => setRunModal(action)}
                />
              </div>
              <form
                className="manual-discussion"
                onSubmit={async (event) => {
                  event.preventDefault();
                  if (!manualText.trim() || manualBusy) return;
                  setManualBusy(true);
                  try {
                    await api.addUtterance(selectedSession.id, {
                      text: manualText.trim(),
                      atMs: Date.now() - Date.parse(selectedSession.startedAt),
                    });
                    setManualText("");
                    await refresh(true);
                  } catch (error) {
                    showToast(
                      error instanceof Error ? error.message : "记录失败",
                    );
                  } finally {
                    setManualBusy(false);
                  }
                }}
              >
                <textarea
                  aria-label="补充讨论"
                  placeholder="也可以打字补充讨论…"
                  maxLength={4000}
                  rows={2}
                  value={manualText}
                  onChange={(event) => setManualText(event.target.value)}
                />
                <button disabled={manualBusy || !manualText.trim()}>
                  记录
                </button>
                <small>
                  录音保存在目标电脑。实时转写由浏览器提供，可能使用其在线识别服务。
                </small>
              </form>
            </>
          )}
        </main>
      </div>
      {systemModal && (
        <SystemSettingsModal
          settings={state.settings}
          onClose={() => setSystemModal(false)}
          onSave={async (values) => {
            try {
              await api.updateSettings(values);
              setSystemModal(false);
              await refresh(true);
              showToast("系统设置已保存");
            } catch (error) {
              showToast(
                error instanceof Error ? error.message : "保存设置失败",
              );
            }
          }}
        />
      )}
      {projectModal && (
        <ProjectModal
          project={projectModal === "new" ? undefined : projectModal}
          onClose={() => setProjectModal(null)}
          onSave={async (values) => {
            try {
              const project =
                projectModal === "new"
                  ? await api.createProject(values)
                  : await api.updateProject(projectModal.id, values);
              setSelectedProjectId(project.id);
              setProjectModal(null);
              await refresh(true);
            } catch (error) {
              showToast(
                error instanceof Error ? error.message : "保存项目失败",
              );
            }
          }}
        />
      )}
      {renameSession && (
        <RenameSessionModal
          session={renameSession}
          onClose={() => setRenameSession(null)}
          onSave={async (title) => {
            try {
              await api.updateSession(renameSession.id, { title });
              setRenameSession(null);
              await refresh(true);
            } catch (error) {
              showToast(error instanceof Error ? error.message : "重命名失败");
            }
          }}
        />
      )}
      {actionModal && selectedSession && (
        <ActionModal
          action={actionModal === "new" ? undefined : actionModal}
          onClose={() => setActionModal(null)}
          onSave={async (values) => {
            await mutateAction(() =>
              actionModal === "new"
                ? api.createAction(selectedSession.id, values)
                : api.updateAction(actionModal.id, values),
            );
            setActionModal(null);
          }}
        />
      )}
      {runModal && (
        <RunModal
          action={runModal}
          state={state}
          onClose={() => setRunModal(null)}
        />
      )}
      {recordingsModal && (
        <RecordingsModal
          session={
            state.sessions.find((item) => item.id === recordingsModal.id) ||
            recordingsModal
          }
          onClose={() => setRecordingsModal(null)}
        />
      )}
      {toast && (
        <div className="toast">
          <CheckCircle2 size={16} /> {toast}
        </div>
      )}
    </div>
  );
}

function EmptyWorkspace({
  project,
  onStart,
  onProject,
}: {
  project?: Project;
  onStart: () => void;
  onProject: () => void;
}) {
  return (
    <div className="empty-workspace">
      <div className="empty-orbit">
        <span>
          <AudioLines size={28} />
        </span>
        <i />
        <i />
        <i />
      </div>
      <p className="overline">CONVERSATION → ACTION</p>
      <h1>
        让讨论自然地
        <br />
        变成下一步。
      </h1>
      <p className="empty-lead">
        新建一个暂停中的会话，准备好后再继续录音。
        <br />
        SaySo 会持续转写、分析，并把明确的事交给 Codex。
      </p>
      <div className="empty-actions">
        <button className="primary-button" onClick={onStart}>
          <Plus size={17} /> 新建会话
        </button>
        {!project && (
          <button className="ghost-button" onClick={onProject}>
            <FolderGit2 size={16} /> 添加项目
          </button>
        )}
      </div>
    </div>
  );
}

function HighlightedDiscussionText({
  text,
  query,
}: {
  text: string;
  query: string;
}) {
  return (
    <>
      {splitDiscussionSearchMatches(text, query).map((part, index) =>
        part.match ? (
          <mark
            className="discussion-search-highlight"
            key={`${index}-${part.text}`}
          >
            {part.text}
          </mark>
        ) : (
          <span key={`${index}-${part.text}`}>{part.text}</span>
        ),
      )}
    </>
  );
}

function TranscriptPanel({
  session,
  interimText,
  isRecording,
  search,
  setSearch,
}: {
  session: Session;
  interimText: string;
  isRecording: boolean;
  search: string;
  setSearch: (value: string) => void;
}) {
  const endRef = useRef<HTMLDivElement>(null);
  const utteranceRefs = useRef(new Map<string, HTMLElement>());
  const normalizedSearch = normalizeDiscussionSearch(search);
  const matches = useMemo(
    () =>
      session.transcript.filter((item) =>
        textMatchesDiscussionSearch(item.text, normalizedSearch),
      ),
    [session.transcript, normalizedSearch],
  );
  const matchIndexById = useMemo(
    () => new Map(matches.map((item, index) => [item.id, index])),
    [matches],
  );
  const matchSignature = matches.map((item) => item.id).join("|");
  const [currentMatchIndex, setCurrentMatchIndex] = useState(-1);

  useEffect(() => {
    setCurrentMatchIndex(normalizedSearch && matches.length ? 0 : -1);
  }, [normalizedSearch, session.id]);
  useEffect(() => {
    setCurrentMatchIndex((current) =>
      matches.length === 0
        ? -1
        : Math.min(Math.max(current, 0), matches.length - 1),
    );
  }, [matches.length]);
  useEffect(() => {
    if (currentMatchIndex < 0) return;
    const target = utteranceRefs.current.get(matches[currentMatchIndex]?.id);
    target?.scrollIntoView({ behavior: "smooth", block: "center" });
  }, [currentMatchIndex, matchSignature, normalizedSearch]);
  useEffect(() => {
    if (!normalizedSearch)
      endRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [session.transcript.length, interimText, normalizedSearch]);

  const moveMatch = (direction: -1 | 1) => {
    if (!matches.length) return;
    setCurrentMatchIndex((current) =>
      current < 0
        ? direction === 1
          ? 0
          : matches.length - 1
        : (current + direction + matches.length) % matches.length,
    );
  };
  return (
    <section className="transcript-panel">
      <div className="panel-header">
        <div>
          <h2>讨论记录</h2>
          <span>{session.transcript.length} 条</span>
        </div>
        <div className="transcript-tools">
          <div
            className={`discussion-search ${normalizedSearch ? "has-query" : ""}`}
          >
            <label className="compact-search">
              <Search size={14} />
              <input
                aria-label="搜索讨论"
                value={search}
                onChange={(event) => setSearch(event.target.value)}
                onKeyDown={(event) => {
                  if (event.key === "Enter") {
                    event.preventDefault();
                    moveMatch(event.shiftKey ? -1 : 1);
                  }
                  if (event.key === "Escape") setSearch("");
                }}
                placeholder="搜索讨论"
              />
            </label>
            {normalizedSearch && (
              <div className="search-navigation">
                <span>
                  {matches.length
                    ? `${currentMatchIndex + 1}/${matches.length}`
                    : "0/0"}
                </span>
                <button
                  type="button"
                  disabled={!matches.length}
                  onClick={() => moveMatch(-1)}
                  aria-label="上一个搜索结果"
                >
                  <ChevronUp size={13} />
                </button>
                <button
                  type="button"
                  disabled={!matches.length}
                  onClick={() => moveMatch(1)}
                  aria-label="下一个搜索结果"
                >
                  <ChevronDown size={13} />
                </button>
              </div>
            )}
          </div>
          <span className={`live-chip ${isRecording ? "on" : ""}`}>
            <i /> {isRecording ? "录音中" : "已暂停"}
          </span>
        </div>
      </div>
      <div className="transcript-scroll">
        {session.transcript.length === 0 && !interimText ? (
          <div className="transcript-empty">
            <div className={`sound-wave ${isRecording ? "active" : ""}`}>
              <i />
              <i />
              <i />
              <i />
              <i />
            </div>
            <h3>{isRecording ? "正在听…" : "讨论已暂停"}</h3>
            <p>
              {isRecording
                ? "识别出的讨论会自动出现在这里。"
                : "点击右上角“继续录音”开始记录。"}
            </p>
          </div>
        ) : (
          <div className="utterance-list">
            {session.transcript.map((utterance, index) => {
              const minute = systemMinuteKey(utterance.createdAt);
              const startsMinute =
                index === 0 ||
                minute !==
                  systemMinuteKey(session.transcript[index - 1].createdAt);
              const timeLabel = formatSystemMinute(utterance.createdAt);
              const matchIndex = matchIndexById.get(utterance.id);
              const isCurrentMatch =
                matchIndex !== undefined && matchIndex === currentMatchIndex;
              return (
                <article
                  className={`utterance ${matchIndex !== undefined ? "discussion-search-match" : ""} ${isCurrentMatch ? "current-search-match" : ""}`}
                  key={utterance.id}
                  ref={(element) => {
                    if (element)
                      utteranceRefs.current.set(utterance.id, element);
                    else utteranceRefs.current.delete(utterance.id);
                  }}
                >
                  <time
                    className={startsMinute ? "minute-mark" : undefined}
                    dateTime={utterance.createdAt}
                    aria-label={
                      startsMinute ? `系统时间 ${timeLabel}` : undefined
                    }
                    aria-hidden={!startsMinute || undefined}
                  >
                    {startsMinute ? timeLabel : null}
                  </time>
                  <p>
                    <HighlightedDiscussionText
                      text={utterance.text}
                      query={normalizedSearch}
                    />
                  </p>
                </article>
              );
            })}
            {interimText && (
              <article className="utterance interim">
                <time>
                  <Radio size={10} /> LIVE
                </time>
                <p>
                  {interimText}
                  <span className="typing-cursor" />
                </p>
              </article>
            )}
            <div ref={endRef} />
          </div>
        )}
      </div>
    </section>
  );
}

function ActionPanel({
  session,
  actions,
  runs,
  settings,
  onCreate,
  onExtract,
  onEdit,
  onDelete,
  onReady,
  onDispatch,
  onViewRun,
}: {
  session: Session;
  actions: ActionItem[];
  runs: AppState["runs"];
  settings: AppSettings;
  onCreate: () => void;
  onExtract: () => void;
  onEdit: (action: ActionItem) => void;
  onDelete: (action: ActionItem) => void;
  onReady: (action: ActionItem) => void;
  onDispatch: (action: ActionItem) => void;
  onViewRun: (action: ActionItem) => void;
}) {
  const analyzing = ["queued", "running"].includes(
    session.analysisStatus || "",
  );
  const [analysisExpanded, setAnalysisExpanded] = useState(true);
  const [detailsOpen, setDetailsOpen] = useState(false);
  const activeTrace = session.analysisTraces?.find(
    (trace) => trace.id === session.analysisActiveTraceId,
  );
  const latestTrace = activeTrace || session.analysisTraces?.[0];
  const previousTraceRef = useRef<string | undefined>(undefined);
  useEffect(() => {
    if (
      analyzing &&
      latestTrace?.id &&
      previousTraceRef.current !== latestTrace.id
    ) {
      previousTraceRef.current = latestTrace.id;
      setAnalysisExpanded(true);
    }
  }, [analyzing, latestTrace?.id]);
  const defaultSteps = [
    "读取最近 5 分钟的原始讨论",
    `对照 ${actions.filter((action) => !["running", "done"].includes(action.status)).length} 条已有需求`,
    `${settings.analysisModel === "default" ? "Codex" : settings.analysisModel} 正在判断补充与修正`,
  ];
  const analysisSteps =
    latestTrace && latestTrace.events.length >= 3
      ? latestTrace.events.slice(-3)
      : defaultSteps;
  const typedSteps = useTypewriterLines(
    analysisSteps,
    analyzing && analysisExpanded,
  );
  const currentStep = [...typedSteps].reverse().find(Boolean);
  return (
    <aside className="action-panel">
      <div className="action-heading">
        <div>
          <span className="action-kicker">
            <Zap size={13} fill="currentColor" /> ACTIONS
          </span>
          <h2>接下来要做的事</h2>
        </div>
        <button
          className="icon-button bordered"
          onClick={onCreate}
          aria-label="添加 Action"
        >
          <Plus size={18} />
        </button>
      </div>
      <div
        className={`analysis-strip ${session.analysisStatus === "failed" ? "failed" : ""} ${analysisExpanded ? "expanded" : ""}`}
      >
        {analyzing ? (
          <LoaderCircle size={13} className="spin" />
        ) : (
          <Sparkles size={13} />
        )}
        <button
          className="analysis-status-copy"
          onClick={() => analyzing && setAnalysisExpanded((value) => !value)}
          disabled={!analyzing}
        >
          <span>
            {analyzing
              ? analysisExpanded
                ? currentStep || "分析中"
                : "分析中"
              : session.analysisStatus === "failed"
                ? session.analysisError || "分析失败，点击重试"
                : session.status === "recording"
                  ? "每 2 分钟分析最近 5 分钟讨论"
                  : `${actions.length} 条 Action`}
          </span>
          {analyzing &&
            (analysisExpanded ? (
              <ChevronDown size={12} />
            ) : (
              <ChevronRight size={12} />
            ))}
        </button>
        {latestTrace && (
          <button
            className="analysis-details-button"
            onClick={() => setDetailsOpen(true)}
          >
            详情
          </button>
        )}
        <button
          className="analysis-refresh"
          onClick={onExtract}
          title="重新分析讨论"
        >
          <RefreshCw size={13} />
        </button>
        {analyzing && analysisExpanded && (
          <div className="analysis-detail-lines">
            {analysisSteps.map((step, index) => (
              <div
                className={
                  typedSteps[index] === step
                    ? "complete"
                    : typedSteps[index]
                      ? "typing"
                      : ""
                }
                key={`${index}-${step}`}
              >
                <span>{index + 1}</span>
                <p>{typedSteps[index] || "\u00a0"}</p>
              </div>
            ))}
          </div>
        )}
      </div>
      <div className="action-list">
        {actions.length === 0 ? (
          <div className="action-empty-compact">
            <Sparkles size={17} />
            <span>{analyzing ? "正在整理讨论…" : "还没有明确的 Action"}</span>
          </div>
        ) : (
          actions.map((action) => {
            const run = runs.find((item) => item.id === action.runId);
            return (
              <article
                className={`action-card card-${action.status}`}
                key={action.id}
              >
                <div className="action-card-top">
                  <StatusPill status={action.status} />
                  <span className={`priority priority-${action.priority}`}>
                    {action.priority === "high"
                      ? "高优先级"
                      : action.priority === "low"
                        ? "低优先级"
                        : "普通"}
                  </span>
                  <div className="card-tools">
                    {!["running", "done"].includes(action.status) && (
                      <button onClick={() => onEdit(action)} aria-label="编辑">
                        <PencilLine size={14} />
                      </button>
                    )}
                    {!["running", "done"].includes(action.status) && (
                      <button
                        onClick={() => onDelete(action)}
                        aria-label="删除"
                      >
                        <Trash2 size={14} />
                      </button>
                    )}
                  </div>
                </div>
                <RequirementAnalysisTicker
                  action={action}
                  trace={latestTrace}
                  active={
                    analyzing && !["running", "done"].includes(action.status)
                  }
                />
                <h3>{action.title}</h3>
                {action.detail && (
                  <p className="action-detail">{action.detail}</p>
                )}
                <blockquote>
                  <span>“</span>
                  {action.source}
                </blockquote>
                {action.status === "running" && (
                  <button
                    className="run-progress"
                    onClick={() => onViewRun(action)}
                  >
                    <span>
                      <LoaderCircle size={15} /> Codex 正在项目中工作
                    </span>
                    <small>查看输出</small>
                  </button>
                )}
                {action.status === "done" && (
                  <button
                    className="run-complete"
                    onClick={() => onViewRun(action)}
                  >
                    <span>
                      <CheckCircle2 size={15} /> 已由 Codex 完成
                    </span>
                    <small>查看结果</small>
                  </button>
                )}
                {action.status === "failed" && run?.error && (
                  <p className="run-error">{run.error}</p>
                )}
                {!["running", "done"].includes(action.status) && (
                  <div className="action-card-footer">
                    <button
                      className={`ready-toggle ${action.status === "ready" ? "checked" : ""}`}
                      onClick={() => onReady(action)}
                    >
                      <span>
                        {action.status === "ready" && <Check size={11} />}
                      </span>{" "}
                      讨论完善
                    </button>
                    <button
                      className="dispatch-button"
                      disabled={action.status === "draft"}
                      onClick={() => onDispatch(action)}
                    >
                      <Send size={14} />{" "}
                      {action.status === "failed"
                        ? "重新交给 Codex"
                        : "交给 Codex"}
                    </button>
                  </div>
                )}
              </article>
            );
          })
        )}
      </div>
      <div className="action-footnote">
        <Code2 size={13} />
        <span>Codex 将在当前项目目录内执行</span>
      </div>
      {detailsOpen && latestTrace && (
        <AnalysisDetailsModal
          trace={latestTrace}
          onClose={() => setDetailsOpen(false)}
        />
      )}
    </aside>
  );
}

function RequirementAnalysisTicker({
  action,
  trace,
  active,
}: {
  action: ActionItem;
  trace?: AnalysisTrace;
  active: boolean;
}) {
  const [expanded, setExpanded] = useState(true);
  const previousTraceRef = useRef<string | undefined>(undefined);
  useEffect(() => {
    if (active && trace?.id && previousTraceRef.current !== trace.id) {
      previousTraceRef.current = trace.id;
      setExpanded(true);
    }
  }, [active, trace?.id]);
  const lines = [
    `输入：核对${trace?.contextLabel || "最近 5 分钟"}的 ${trace?.utteranceCount || 0} 条讨论`,
    `判断“${action.title}”是否需要补充或修正`,
  ];
  const typedLines = useTypewriterLines(lines, active && expanded);
  if (!active) return null;
  return (
    <div className={`requirement-analysis ${expanded ? "expanded" : ""}`}>
      <button
        className="requirement-analysis-toggle"
        onClick={() => setExpanded((value) => !value)}
      >
        <LoaderCircle size={11} className="spin" />
        <span>{expanded ? "需求分析中" : "分析中"}</span>
        {expanded ? <ChevronDown size={11} /> : <ChevronRight size={11} />}
      </button>
      {expanded && (
        <div className="requirement-analysis-lines">
          {lines.map((line, index) => (
            <p
              className={
                typedLines[index] === line
                  ? "complete"
                  : typedLines[index]
                    ? "typing"
                    : ""
              }
              key={line}
            >
              {typedLines[index] || "\u00a0"}
            </p>
          ))}
        </div>
      )}
    </div>
  );
}

function AnalysisDetailsModal({
  trace,
  onClose,
}: {
  trace: AnalysisTrace;
  onClose: () => void;
}) {
  const liveEndRef = useRef<HTMLDivElement>(null);
  useEffect(() => {
    liveEndRef.current?.scrollIntoView({
      behavior: "smooth",
      block: "nearest",
    });
  }, [trace.events.length, trace.output]);
  const statusLabel = {
    queued: "排队中",
    running: "分析中",
    done: "已完成",
    failed: "失败",
  }[trace.status];
  return (
    <Modal
      wide
      title="分析详情"
      description={`${trace.model} · ${statusLabel} · ${trace.utteranceCount} 条原始讨论`}
      onClose={onClose}
    >
      <div className="analysis-trace-scroll">
        <section className="analysis-trace-section">
          <div className="analysis-trace-heading">
            <h3>实时过程</h3>
            <span className={`trace-status trace-${trace.status}`}>
              {trace.status === "running" && (
                <LoaderCircle size={11} className="spin" />
              )}
              {statusLabel}
            </span>
          </div>
          <div className="analysis-live-log">
            {trace.events.map((event, index) => (
              <div key={`${index}-${event}`}>
                <span>{String(index + 1).padStart(2, "0")}</span>
                <p>{event}</p>
              </div>
            ))}
            <div ref={liveEndRef} />
          </div>
        </section>
        <section className="analysis-trace-section">
          <h3>原始输入</h3>
          <pre>{trace.input || "暂无输入"}</pre>
        </section>
        <section className="analysis-trace-section">
          <h3>最终输出</h3>
          <pre>
            {trace.output ||
              (trace.status === "running"
                ? "等待模型返回最终结构化输出…"
                : trace.error || "本轮没有输出")}
          </pre>
        </section>
      </div>
      <div className="modal-actions">
        <button className="primary-button" onClick={onClose}>
          关闭
        </button>
      </div>
    </Modal>
  );
}

function SystemSettingsModal({
  settings,
  onClose,
  onSave,
}: {
  settings: AppSettings;
  onClose: () => void;
  onSave: (values: AppSettings) => Promise<void>;
}) {
  const [values, setValues] = useState(settings);
  const [saving, setSaving] = useState(false);
  return (
    <Modal
      title="系统设置"
      description="配置讨论如何被分析，以及 Action 由谁执行。"
      onClose={onClose}
    >
      <form
        onSubmit={(event) => {
          event.preventDefault();
          setSaving(true);
          void onSave(values).finally(() => setSaving(false));
        }}
      >
        <div className="setting-group">
          <div className="setting-copy">
            <strong>讨论分析</strong>
            <small>从转写中持续提炼决定和待办</small>
          </div>
          <select
            value={values.analysisProvider}
            onChange={(event) =>
              setValues({
                ...values,
                analysisProvider: event.target
                  .value as AppSettings["analysisProvider"],
              })
            }
          >
            <option value="codex">Codex</option>
            <option value="local">本地规则</option>
          </select>
        </div>
        <div className="setting-group">
          <div className="setting-copy">
            <strong>分析模型</strong>
            <small>小模型更快，也更节省额度</small>
          </div>
          <select
            disabled={values.analysisProvider === "local"}
            value={values.analysisModel}
            onChange={(event) =>
              setValues({ ...values, analysisModel: event.target.value })
            }
          >
            {modelOptions.map((model) => (
              <option value={model.value} key={model.value}>
                {model.label}
              </option>
            ))}
          </select>
        </div>
        <div className="setting-group">
          <div className="setting-copy">
            <strong>持续分析</strong>
            <small>录音中每 2 分钟运行一次，每次读取最近 5 分钟讨论</small>
          </div>
          <span className="setting-fixed-value">2 分钟 / 5 分钟窗口</span>
        </div>
        <div className="setting-group">
          <div className="setting-copy">
            <strong>执行模型</strong>
            <small>最终进入项目实现 Action 时使用</small>
          </div>
          <select
            value={values.executionModel}
            onChange={(event) =>
              setValues({ ...values, executionModel: event.target.value })
            }
          >
            {executionModelOptions.map((model) => (
              <option value={model.value} key={model.value}>
                {model.label}
              </option>
            ))}
          </select>
        </div>
        <div className="setting-group">
          <div className="setting-copy">
            <strong>Action 执行器</strong>
            <small>在项目目录内完成已确认的工作</small>
          </div>
          <select value="codex" disabled>
            <option value="codex">Codex</option>
          </select>
        </div>
        <div className="modal-actions">
          <button type="button" className="ghost-button" onClick={onClose}>
            取消
          </button>
          <button className="primary-button" disabled={saving}>
            {saving ? <LoaderCircle size={15} /> : <Check size={15} />} 保存设置
          </button>
        </div>
      </form>
    </Modal>
  );
}

function RecordingRow({
  recording,
  index,
}: {
  recording: RecordingSegment;
  index: number;
}) {
  const [audioURL, setAudioURL] = useState("");
  const [audioError, setAudioError] = useState("");
  const [audioLoading, setAudioLoading] = useState(false);
  const controllerRef = useRef<AbortController | null>(null);
  useEffect(
    () => () => {
      controllerRef.current?.abort();
    },
    [],
  );
  useEffect(
    () => () => {
      if (audioURL) URL.revokeObjectURL(audioURL);
    },
    [audioURL],
  );
  const loadAudio = async () => {
    setAudioLoading(true);
    setAudioError("");
    const controller = new AbortController();
    controllerRef.current = controller;
    try {
      setAudioURL(await recordingURL(recording.id, controller.signal));
    } catch (error) {
      if (!controller.signal.aborted)
        setAudioError(error instanceof Error ? error.message : "无法读取录音");
    } finally {
      if (!controller.signal.aborted) setAudioLoading(false);
    }
  };
  const [detectedDuration, setDetectedDuration] = useState<
    number | undefined
  >();
  const duration = recording.durationMs ?? detectedDuration;
  const detectLegacyDuration = (audio: HTMLAudioElement) => {
    if (Number.isFinite(audio.duration) && audio.duration > 0) {
      setDetectedDuration(audio.duration * 1000);
      if (audio.dataset.durationProbe === "active") {
        audio.currentTime = 0;
        delete audio.dataset.durationProbe;
      }
      return;
    }
    if (
      recording.durationMs === undefined &&
      audio.dataset.durationProbe !== "active"
    ) {
      audio.dataset.durationProbe = "active";
      audio.currentTime = Number.MAX_SAFE_INTEGER;
    }
  };
  return (
    <article className="recording-row">
      <div className="recording-index">
        <FileAudio size={15} />
        <span>{String(index + 1).padStart(2, "0")}</span>
      </div>
      <div className="recording-info">
        <div>
          <strong>录音片段 {index + 1}</strong>
          <span>
            {new Date(recording.startedAt).toLocaleString("zh-CN", {
              month: "numeric",
              day: "numeric",
              hour: "2-digit",
              minute: "2-digit",
            })}
          </span>
        </div>
        <small>
          {duration !== undefined ? formatClock(duration) : "读取时长中"} ·{" "}
          {formatBytes(recording.sizeBytes)}
        </small>
        {!audioURL && (
          <button
            className="ghost-button"
            disabled={audioLoading}
            onClick={() => void loadAudio()}
          >
            {audioLoading ? "正在从设备读取…" : "加载录音"}
          </button>
        )}
        {audioError && <p role="alert">{audioError}</p>}
        {audioURL && (
          <audio
            controls
            preload="metadata"
            src={audioURL}
            onLoadedMetadata={(event) =>
              detectLegacyDuration(event.currentTarget)
            }
            onDurationChange={(event) =>
              detectLegacyDuration(event.currentTarget)
            }
            onSeeked={(event) => detectLegacyDuration(event.currentTarget)}
            onPlay={(event) => {
              (event.currentTarget.getRootNode() as ShadowRoot)
                .querySelectorAll("audio")
                .forEach((audio) => {
                  if (audio !== event.currentTarget) audio.pause();
                });
            }}
          />
        )}
      </div>
    </article>
  );
}

function RecordingsModal({
  session,
  onClose,
}: {
  session: Session;
  onClose: () => void;
}) {
  const recordings = [...(session.recordings || [])].sort((a, b) =>
    a.startedAt.localeCompare(b.startedAt),
  );
  return (
    <Modal
      wide
      title="录音片段"
      description={`${session.title} · 共 ${recordings.length} 段，每段最长 10 分钟`}
      onClose={onClose}
    >
      <div className="recordings-list">
        {recordings.length ? (
          recordings.map((recording, index) => (
            <RecordingRow
              recording={recording}
              index={index}
              key={recording.id}
            />
          ))
        ) : (
          <div className="recordings-empty">
            <FileAudio size={20} />
            暂时没有录音
          </div>
        )}
      </div>
      <div className="modal-actions">
        <button className="primary-button" onClick={onClose}>
          关闭
        </button>
      </div>
    </Modal>
  );
}

function ProjectModal({
  project,
  onClose,
  onSave,
}: {
  project?: Project;
  onClose: () => void;
  onSave: (values: Pick<Project, "name" | "rootPath">) => Promise<void>;
}) {
  const [name, setName] = useState(project?.name || "");
  const [rootPath, setRootPath] = useState(project?.rootPath || "");
  const [saving, setSaving] = useState(false);
  return (
    <Modal
      title={project ? "编辑项目" : "添加项目"}
      description="Action 会在这个本地目录中交给 Codex。"
      onClose={onClose}
    >
      <form
        onSubmit={(event) => {
          event.preventDefault();
          setSaving(true);
          void onSave({ name, rootPath }).finally(() => setSaving(false));
        }}
      >
        <label className="field-label">项目名称</label>
        <input
          className="text-field"
          value={name}
          onChange={(event) => setName(event.target.value)}
          placeholder="我的产品"
        />
        <label className="field-label">本地项目目录</label>
        <div className="path-field">
          <FolderGit2 size={16} />
          <input
            value={rootPath}
            onChange={(event) => setRootPath(event.target.value)}
            placeholder="C:\\Code\\my-project"
          />
        </div>
        <div className="modal-actions">
          <button type="button" className="ghost-button" onClick={onClose}>
            取消
          </button>
          <button
            className="primary-button"
            disabled={saving || !name.trim() || !rootPath.trim()}
          >
            {saving ? <LoaderCircle size={15} /> : <Check size={15} />} 保存项目
          </button>
        </div>
      </form>
    </Modal>
  );
}

function RenameSessionModal({
  session,
  onClose,
  onSave,
}: {
  session: Session;
  onClose: () => void;
  onSave: (title: string) => Promise<void>;
}) {
  const [title, setTitle] = useState(session.title);
  const [saving, setSaving] = useState(false);
  return (
    <Modal
      title="重命名会话"
      description="日期时间名称也可以随时改掉。"
      onClose={onClose}
    >
      <form
        onSubmit={(event) => {
          event.preventDefault();
          setSaving(true);
          void onSave(title).finally(() => setSaving(false));
        }}
      >
        <label className="field-label">会话名称</label>
        <input
          className="text-field"
          autoFocus
          value={title}
          onChange={(event) => setTitle(event.target.value)}
        />
        <div className="modal-actions">
          <button type="button" className="ghost-button" onClick={onClose}>
            取消
          </button>
          <button className="primary-button" disabled={saving || !title.trim()}>
            {saving ? <LoaderCircle size={15} /> : <Check size={15} />} 保存
          </button>
        </div>
      </form>
    </Modal>
  );
}

function ActionModal({
  action,
  onClose,
  onSave,
}: {
  action?: ActionItem;
  onClose: () => void;
  onSave: (
    values: Pick<ActionItem, "title" | "detail" | "priority">,
  ) => Promise<void>;
}) {
  const [title, setTitle] = useState(action?.title || "");
  const [detail, setDetail] = useState(action?.detail || "");
  const [priority, setPriority] = useState<ActionPriority>(
    action?.priority || "medium",
  );
  const [saving, setSaving] = useState(false);
  return (
    <Modal
      title={action ? "完善 Action" : "添加 Action"}
      description="把目标和验收结果说清楚，Codex 会做得更稳。"
      onClose={onClose}
    >
      <form
        onSubmit={(event: FormEvent) => {
          event.preventDefault();
          setSaving(true);
          void onSave({ title, detail, priority }).finally(() =>
            setSaving(false),
          );
        }}
      >
        <label className="field-label">要做什么</label>
        <input
          className="text-field"
          value={title}
          autoFocus
          onChange={(event) => setTitle(event.target.value)}
          placeholder="一句话描述结果"
        />
        <label className="field-label">详细说明</label>
        <textarea
          className="text-field textarea"
          value={detail}
          onChange={(event) => setDetail(event.target.value)}
          placeholder="补充范围、约束和验收标准…"
        />
        <label className="field-label">优先级</label>
        <div className="priority-picker">
          {(["high", "medium", "low"] as ActionPriority[]).map((value) => (
            <button
              type="button"
              className={priority === value ? "active" : ""}
              key={value}
              onClick={() => setPriority(value)}
            >
              <Circle size={9} fill="currentColor" />{" "}
              {value === "high" ? "高" : value === "medium" ? "普通" : "低"}
            </button>
          ))}
        </div>
        <div className="modal-actions">
          <button type="button" className="ghost-button" onClick={onClose}>
            取消
          </button>
          <button className="primary-button" disabled={saving || !title.trim()}>
            {saving ? <LoaderCircle size={15} /> : <Check size={15} />} 保存
            Action
          </button>
        </div>
      </form>
    </Modal>
  );
}

function RunModal({
  action,
  state,
  onClose,
}: {
  action: ActionItem;
  state: AppState;
  onClose: () => void;
}) {
  const [cancelError, setCancelError] = useState("");
  action = state.actions.find((item) => item.id === action.id) || action;
  const run = state.runs.find((item) => item.id === action.runId);
  return (
    <Modal
      wide
      title={
        action.status === "running"
          ? "Codex 正在工作"
          : action.status === "done"
            ? "Codex 已完成"
            : "Codex 执行结果"
      }
      description={action.title}
      onClose={onClose}
    >
      <div className="run-meta">
        <StatusPill status={action.status} />
        {run && (
          <span>
            <Clock3 size={13} />{" "}
            {new Date(run.startedAt).toLocaleString("zh-CN")}
          </span>
        )}
      </div>
      <pre className="run-output">
        {run?.output ||
          (run?.status === "running"
            ? "正在启动 Codex…"
            : run?.error || "暂无输出")}
      </pre>
      {run?.error && <p role="alert">{run.error}</p>}
      {cancelError && <p role="alert">{cancelError}</p>}
      <div className="modal-actions">
        {run?.status === "running" && (
          <button
            className="ghost-button"
            onClick={() =>
              void api
                .cancelRun(run.id)
                .catch((error) => setCancelError(error.message))
            }
          >
            停止执行
          </button>
        )}
        <button className="primary-button" onClick={onClose}>
          关闭
        </button>
      </div>
    </Modal>
  );
}
