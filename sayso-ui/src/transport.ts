export type Transport = (
  method: string,
  path: string,
  body?: unknown,
) => Promise<any>;
let transport: Transport = async () => {
  throw Error("请从已连接的设备打开 SaySo");
};
export function configure(next: Transport) {
  transport = next;
}
export function call<T>(
  method: string,
  path: string,
  body?: unknown,
): Promise<T> {
  return transport(method, path, body);
}
export function fromBase64(value: string) {
  return Uint8Array.from(atob(value), (c) => c.charCodeAt(0));
}
let uploadQueue: Promise<unknown> = Promise.resolve();
export function upload(
  sessionId: string,
  blob: Blob,
  meta: { startedAt: string; endedAt: string; durationMs: number },
) {
  const transportForDevice = transport;
  const next = uploadQueue
    .catch(() => {})
    .then(async () => {
      const { id } = await transportForDevice("POST", "/api/audio/begin", {
        ...meta,
        sessionId,
        sizeBytes: blob.size,
        mimeType: blob.type || "audio/webm",
      });
      try {
        for (let offset = 0; offset < blob.size; offset += 16000) {
          const bytes = new Uint8Array(
            await blob.slice(offset, offset + 16000).arrayBuffer(),
          );
          await transportForDevice("POST", `/api/audio/${id}/chunk`, {
            offset,
            data: btoa(String.fromCharCode(...bytes)),
          });
        }
        return await transportForDevice("POST", `/api/audio/${id}/finish`);
      } catch (error) {
        await transportForDevice("POST", `/api/audio/${id}/abort`).catch(
          () => {},
        );
        throw error;
      }
    });
  uploadQueue = next;
  return next;
}
export async function recordingURL(id: string, signal: AbortSignal) {
  const transportForDevice = transport;
  const chunks: Uint8Array<ArrayBuffer>[] = [];
  let offset = 0,
    size = 0,
    mimeType = "";
  do {
    if (signal.aborted) throw Error("已取消");
    const data = await transportForDevice("POST", `/api/audio/${id}/read`, {
      offset,
    });
    if (!data.size || data.size > 64 * 1024 * 1024)
      throw Error("录音大小不合法");
    const chunk = fromBase64(data.data);
    if (!chunk.length) throw Error("录音数据不完整");
    chunks.push(chunk);
    offset += chunk.length;
    size = data.size;
    mimeType = data.mimeType;
  } while (offset < size);
  if (signal.aborted) throw Error("已取消");
  return URL.createObjectURL(new Blob(chunks, { type: mimeType }));
}
