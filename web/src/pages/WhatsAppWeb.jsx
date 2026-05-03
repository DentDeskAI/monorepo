import { useEffect, useMemo, useRef, useState } from "react";
import clsx from "clsx";
import { waWeb } from "../api/client.js";

function formatChatTime(timestamp) {
  if (!timestamp) return "";
  return new Intl.DateTimeFormat("ru-RU", {
    day: "2-digit",
    month: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  }).format(new Date(timestamp * 1000));
}

function formatMessageTime(message) {
  const date = message.createdAt ? new Date(message.createdAt) : new Date();
  return new Intl.DateTimeFormat("ru-RU", {
    day: "2-digit",
    month: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  }).format(date);
}

function appendUniqueMessage(list, message) {
  if (!message?.id) return list;
  if (list.some((item) => item.id === message.id)) return list;
  return [...list, message].sort((a, b) => (a.timestamp || 0) - (b.timestamp || 0));
}

export default function WhatsAppWeb() {
  const [status, setStatus] = useState(null);
  const [qr, setQr] = useState(null);
  const [chats, setChats] = useState([]);
  const [activeChatId, setActiveChatId] = useState("");
  const [messages, setMessages] = useState([]);
  const [input, setInput] = useState("");
  const [loadingChats, setLoadingChats] = useState(false);
  const [loadingMessages, setLoadingMessages] = useState(false);
  const [sending, setSending] = useState(false);
  const [error, setError] = useState("");

  const activeChatIdRef = useRef(activeChatId);
  const messagesEndRef = useRef(null);

  const activeChat = useMemo(
    () => chats.find((chat) => chat.id === activeChatId) || null,
    [activeChatId, chats],
  );

  async function refreshStatus() {
    try {
      const nextStatus = await waWeb.status();
      setStatus(nextStatus);
      if (nextStatus.hasQr && !nextStatus.ready) {
        const response = await waWeb.qr();
        setQr(response.qr || null);
      }
      if (nextStatus.ready) setQr(null);
    } catch (err) {
      setError(err.message);
    }
  }

  async function loadChats() {
    setLoadingChats(true);
    try {
      const data = await waWeb.chats();
      setChats(data.chats || []);
      setError("");
    } catch (err) {
      setError(err.message);
    } finally {
      setLoadingChats(false);
    }
  }

  async function loadMessages(chatId) {
    if (!chatId) {
      setMessages([]);
      return;
    }
    setLoadingMessages(true);
    try {
      const data = await waWeb.messages(chatId);
      setMessages(data.messages || []);
      setError("");
    } catch (err) {
      setError(err.message);
    } finally {
      setLoadingMessages(false);
    }
  }

  async function sendMessage(event) {
    event.preventDefault();
    const body = input.trim();
    if (!activeChatId || !body || sending) return;

    setSending(true);
    try {
      const data = await waWeb.send(activeChatId, body);
      if (data.message) {
        setMessages((current) => appendUniqueMessage(current, data.message));
      }
      setInput("");
      loadChats();
      setError("");
    } catch (err) {
      setError(err.message);
    } finally {
      setSending(false);
    }
  }

  async function logout() {
    try {
      await waWeb.logout();
      setChats([]);
      setMessages([]);
      setActiveChatId("");
      setQr(null);
      refreshStatus();
      setError("");
    } catch (err) {
      setError(err.message);
    }
  }

  async function resetSession() {
    try {
      await waWeb.resetSession();
      setChats([]);
      setMessages([]);
      setActiveChatId("");
      setQr(null);
      refreshStatus();
      setError("");
    } catch (err) {
      setError(err.message);
    }
  }

  useEffect(() => {
    activeChatIdRef.current = activeChatId;
    loadMessages(activeChatId);
  }, [activeChatId]);

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages]);

  useEffect(() => {
    refreshStatus();

    const source = new EventSource(waWeb.eventsUrl());

    source.addEventListener("status", (event) => {
      const nextStatus = JSON.parse(event.data);
      setStatus(nextStatus);
      if (nextStatus.ready) {
        setQr(null);
        loadChats();
      }
    });

    source.addEventListener("qr", (event) => {
      const data = JSON.parse(event.data);
      setQr(data.qr || null);
    });

    source.addEventListener("message", (event) => {
      const data = JSON.parse(event.data);
      if (data.message?.chatId === activeChatIdRef.current) {
        setMessages((current) => appendUniqueMessage(current, data.message));
      }
      loadChats();
    });

    source.addEventListener("chats", () => {
      loadChats();
    });

    source.onerror = () => {
      setError("WhatsApp realtime connection is reconnecting.");
    };

    return () => source.close();
  }, []);

  const ready = Boolean(status?.ready);
  const statusLabel = status?.ready
    ? "Connected"
    : status?.hasQr
      ? "Scan QR"
      : status?.initializing
        ? "Starting"
        : status?.state || "Disconnected";

  return (
    <div className="h-full overflow-y-auto p-6">
      <div className="space-y-4">
        <div className="rounded-3xl border border-slate-200 bg-white p-5 shadow-sm dark:border-slate-700 dark:bg-slate-900">
          <div className="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
            <div>
              <p className="text-sm font-semibold uppercase tracking-[0.24em] text-emerald-600">
                WhatsApp Web
              </p>
              <h1 className="mt-1 text-2xl font-bold text-slate-950 dark:text-slate-50">
                CRM messenger
              </h1>
              <p className="mt-1 text-sm text-slate-500 dark:text-slate-400">
                Scan once, then keep the Docker volume to preserve the session.
              </p>
            </div>
            <div className="flex flex-wrap items-center gap-2">
              <span
                className={clsx(
                  "rounded-full px-3 py-1 text-sm font-semibold",
                  ready
                    ? "bg-emerald-100 text-emerald-700 dark:bg-emerald-500/15 dark:text-emerald-300"
                    : "bg-amber-100 text-amber-700 dark:bg-amber-500/15 dark:text-amber-300",
                )}
              >
                {statusLabel}
              </span>
              <button
                type="button"
                onClick={refreshStatus}
                className="rounded-xl border border-slate-200 px-4 py-2 text-sm font-semibold text-slate-700 hover:bg-slate-50 dark:border-slate-700 dark:text-slate-200 dark:hover:bg-slate-800"
              >
                Refresh
              </button>
              <button
                type="button"
                onClick={logout}
                className="rounded-xl border border-slate-200 px-4 py-2 text-sm font-semibold text-slate-700 hover:bg-slate-50 dark:border-slate-700 dark:text-slate-200 dark:hover:bg-slate-800"
              >
                Logout
              </button>
              <button
                type="button"
                onClick={resetSession}
                className="rounded-xl bg-rose-50 px-4 py-2 text-sm font-semibold text-rose-700 hover:bg-rose-100 dark:bg-rose-500/15 dark:text-rose-300"
              >
                Reset QR
              </button>
            </div>
          </div>

          {status?.loadingPercent ? (
            <div className="mt-4 h-2 overflow-hidden rounded-full bg-slate-100 dark:bg-slate-800">
              <div
                className="h-full rounded-full bg-emerald-500 transition-all"
                style={{ width: `${Math.min(status.loadingPercent, 100)}%` }}
              />
            </div>
          ) : null}

          {status?.error ? (
            <p className="mt-3 rounded-2xl bg-rose-50 px-4 py-3 text-sm text-rose-700 dark:bg-rose-500/15 dark:text-rose-300">
              {status.error}
            </p>
          ) : null}

          {error ? (
            <p className="mt-3 rounded-2xl bg-amber-50 px-4 py-3 text-sm text-amber-700 dark:bg-amber-500/15 dark:text-amber-300">
              {error}
            </p>
          ) : null}
        </div>

        {!ready ? (
          <div className="grid gap-4 lg:grid-cols-[360px_1fr]">
            <div className="rounded-3xl border border-slate-200 bg-white p-6 text-center shadow-sm dark:border-slate-700 dark:bg-slate-900">
              <div className="mx-auto flex h-72 w-72 items-center justify-center rounded-3xl bg-slate-50 dark:bg-slate-800">
                {qr ? (
                  <img src={qr} alt="WhatsApp login QR" className="h-64 w-64 rounded-2xl" />
                ) : (
                  <span className="text-sm text-slate-500">Waiting for QR...</span>
                )}
              </div>
            </div>
            <div className="rounded-3xl border border-slate-200 bg-white p-6 shadow-sm dark:border-slate-700 dark:bg-slate-900">
              <h2 className="text-xl font-bold text-slate-950 dark:text-slate-50">
                Connect phone
              </h2>
              <div className="mt-4 space-y-3 text-sm text-slate-600 dark:text-slate-300">
                <p>Open WhatsApp on the phone that should be connected to the CRM.</p>
                <p>Go to Linked devices, scan this QR code, then wait until status becomes Connected.</p>
                <p>If the QR expires, press Refresh or Reset QR.</p>
              </div>
            </div>
          </div>
        ) : (
          <div className="grid min-h-[680px] overflow-hidden rounded-3xl border border-slate-200 bg-white shadow-sm dark:border-slate-700 dark:bg-slate-900 lg:grid-cols-[360px_1fr]">
            <aside className="border-b border-slate-200 dark:border-slate-700 lg:border-b-0 lg:border-r">
              <div className="flex items-center justify-between border-b border-slate-200 p-4 dark:border-slate-700">
                <div>
                  <h2 className="font-bold text-slate-950 dark:text-slate-50">Chats</h2>
                  <p className="text-xs text-slate-500">
                    {loadingChats ? "Loading..." : `${chats.length} conversations`}
                  </p>
                </div>
                <button
                  type="button"
                  onClick={loadChats}
                  className="rounded-xl border border-slate-200 px-3 py-2 text-sm font-semibold text-slate-700 hover:bg-slate-50 dark:border-slate-700 dark:text-slate-200 dark:hover:bg-slate-800"
                >
                  Reload
                </button>
              </div>
              <div className="max-h-[620px] overflow-y-auto">
                {chats.map((chat) => (
                  <button
                    key={chat.id}
                    type="button"
                    onClick={() => setActiveChatId(chat.id)}
                    className={clsx(
                      "w-full border-b border-slate-100 p-4 text-left transition dark:border-slate-800",
                      activeChatId === chat.id
                        ? "bg-emerald-50 dark:bg-emerald-500/10"
                        : "hover:bg-slate-50 dark:hover:bg-slate-800",
                    )}
                  >
                    <div className="flex items-start justify-between gap-3">
                      <div className="min-w-0">
                        <p className="truncate font-semibold text-slate-950 dark:text-slate-50">
                          {chat.name}
                        </p>
                        <p className="mt-1 truncate text-sm text-slate-500 dark:text-slate-400">
                          {chat.lastMessage?.body || "No messages yet"}
                        </p>
                      </div>
                      <div className="shrink-0 text-right">
                        <p className="text-xs text-slate-400">
                          {formatChatTime(chat.lastActivityTimestamp)}
                        </p>
                        {chat.unreadCount ? (
                          <span className="mt-2 inline-flex min-w-6 justify-center rounded-full bg-emerald-500 px-2 py-0.5 text-xs font-bold text-white">
                            {chat.unreadCount}
                          </span>
                        ) : null}
                      </div>
                    </div>
                  </button>
                ))}
                {!chats.length && !loadingChats ? (
                  <p className="p-5 text-sm text-slate-500">No chats returned by WhatsApp yet.</p>
                ) : null}
              </div>
            </aside>

            <section className="flex min-h-[680px] flex-col">
              {activeChat ? (
                <>
                  <header className="border-b border-slate-200 p-4 dark:border-slate-700">
                    <h2 className="font-bold text-slate-950 dark:text-slate-50">
                      {activeChat.name}
                    </h2>
                    <p className="text-sm text-slate-500">{activeChat.id}</p>
                  </header>

                  <div className="flex-1 space-y-3 overflow-y-auto bg-slate-50 p-5 dark:bg-slate-950/40">
                    {loadingMessages ? (
                      <p className="text-sm text-slate-500">Loading messages...</p>
                    ) : null}
                    {messages.map((message) => (
                      <div
                        key={message.id}
                        className={clsx("flex", message.fromMe ? "justify-end" : "justify-start")}
                      >
                        <div
                          className={clsx(
                            "max-w-[78%] rounded-3xl px-4 py-3 shadow-sm",
                            message.fromMe
                              ? "bg-emerald-600 text-white"
                              : "bg-white text-slate-950 dark:bg-slate-800 dark:text-slate-50",
                          )}
                        >
                          <p className="whitespace-pre-wrap text-sm leading-relaxed">{message.body}</p>
                          <p
                            className={clsx(
                              "mt-2 text-xs",
                              message.fromMe ? "text-emerald-100" : "text-slate-400",
                            )}
                          >
                            {message.fromMe ? "me" : "contact"} - {formatMessageTime(message)}
                          </p>
                        </div>
                      </div>
                    ))}
                    <div ref={messagesEndRef} />
                  </div>

                  <form
                    onSubmit={sendMessage}
                    className="flex gap-3 border-t border-slate-200 p-4 dark:border-slate-700"
                  >
                    <input
                      value={input}
                      onChange={(event) => setInput(event.target.value)}
                      placeholder="Write a message"
                      className="flex-1 rounded-2xl border border-slate-200 bg-white px-4 py-3 text-sm outline-none focus:border-emerald-500 dark:border-slate-700 dark:bg-slate-950 dark:text-slate-50"
                    />
                    <button
                      type="submit"
                      disabled={!input.trim() || sending}
                      className="rounded-2xl bg-emerald-600 px-6 py-3 text-sm font-bold text-white hover:bg-emerald-700 disabled:cursor-not-allowed disabled:opacity-50"
                    >
                      {sending ? "Sending..." : "Send"}
                    </button>
                  </form>
                </>
              ) : (
                <div className="flex flex-1 items-center justify-center p-8 text-center">
                  <div>
                    <p className="text-lg font-bold text-slate-950 dark:text-slate-50">
                      Select a chat
                    </p>
                    <p className="mt-2 text-sm text-slate-500">
                      Messages will update here in real time while the bridge is connected.
                    </p>
                  </div>
                </div>
              )}
            </section>
          </div>
        )}
      </div>
    </div>
  );
}
