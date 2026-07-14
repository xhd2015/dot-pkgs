import { apiEventSource, apiFetch } from './client';

export type OpAction = 'rebase' | 'sync' | 'tag' | 'push' | 'agent-run';

export type CreateOpRequest = {
  action: OpAction;
  target_id?: string;
  label?: string;
};

export type CreateOpResponse = {
  op_id: string;
  action: string;
};

export type LogLineDTO = {
  ts: string;
  level: string;
  message: string;
};

export type OpDoneDTO = {
  ok: boolean;
  summary?: string;
  error?: string;
};

export async function startOp(req: CreateOpRequest): Promise<CreateOpResponse> {
  const res = await apiFetch('/api/wrk/ops', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(req),
  });
  if (!res.ok) {
    let msg = `start op failed (${res.status})`;
    try {
      const j = (await res.json()) as { error?: string };
      if (j.error) msg = j.error;
    } catch {
      /* ignore */
    }
    throw new Error(msg);
  }
  return (await res.json()) as CreateOpResponse;
}

export type StreamHandlers = {
  onLog: (line: LogLineDTO) => void;
  onDone: (done: OpDoneDTO) => void;
  onError?: (err: Error) => void;
};

/** Open SSE stream for op logs. Returns a close function. */
export function streamOpLogs(opId: string, handlers: StreamHandlers): () => void {
  const es = apiEventSource(`/api/wrk/ops/${encodeURIComponent(opId)}/logs`);

  es.addEventListener('log', (ev) => {
    try {
      const data = JSON.parse((ev as MessageEvent).data) as LogLineDTO;
      handlers.onLog(data);
    } catch (e) {
      handlers.onError?.(e instanceof Error ? e : new Error(String(e)));
    }
  });

  es.addEventListener('done', (ev) => {
    try {
      const data = JSON.parse((ev as MessageEvent).data) as OpDoneDTO;
      handlers.onDone(data);
    } catch (e) {
      handlers.onError?.(e instanceof Error ? e : new Error(String(e)));
    } finally {
      es.close();
    }
  });

  es.onerror = () => {
    // EventSource retries by default; close on hard failure after open.
    if (es.readyState === EventSource.CLOSED) {
      handlers.onError?.(new Error('log stream closed'));
    }
  };

  return () => {
    es.close();
  };
}
