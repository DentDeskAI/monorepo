import { useEffect, useMemo, useRef, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { api } from "../api/client";
import { useTranslation } from "../hooks/useTranslation";

// i18n Keys
const i18nKeys = {
  chats: {
    title: "chats.title",
    count: "chats.count",
    empty: "chats.empty",
    no_patient: "common.no_patient",
    select_chat: "chats.select_chat",
    operator: "chats.operator",
    return_to_bot: "chats.return_to_bot",
    placeholder: "chats.placeholder",
    send: "chats.send",
    sending: "chats.sending",
    loading: "chats.loading",
    info_text: "chats.info_text",
  },
  sender: {
    operator: "chats.sender.operator",
    aigul: "chats.sender.aigul",
    patient: "chats.sender.patient",
  },
};

export default function Chats() {
  const { t } = useTranslation();
  const { id } = useParams();
  const nav = useNavigate();
  const [chats, setChats] = useState([]);
  const [messages, setMessages] = useState([]);
  const [loading, setLoading] = useState(false);
  const [sending, setSending] = useState(false);
  const [input, setInput] = useState("");
  const scrollRef = useRef(null);

  const loadChats = () => api.chats().then(setChats).catch(() => {});
  useEffect(() => { loadChats(); }, []);

  useEffect(() => {
    if (!id) { setMessages([]); return; }
    setLoading(true);
    api.messages(id).then((msgs) => setMessages(msgs || [])).finally(() => setLoading(false));
  }, [id]);

  useEffect(() => {
    const unsub = api.subscribe((type, payload) => {
      if (type === "message") {
        loadChats();
        if (id && payload.conversation_id === id) {
          setMessages((prev) => {
            if (prev.some((m) => m.id === payload.id)) return prev;
            return [...prev, payload];
          });
        }
      }
    });
    return unsub;
  }, [id]);

  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  }, [messages]);

  const current = useMemo(
      () => chats.find((c) => c.conversation.id === id),
      [chats, id]
  );

  const send = async () => {
    if (!input.trim() || !id) return;
    setSending(true);
    try {
      const msg = await api.send(id, input.trim());
      setMessages((prev) => [...prev, msg]);
      setInput("");
      loadChats();
    } catch (e) {
      alert(`${t("chats.send_failed")}: ${e.message}`);
    } finally {
      setSending(false);
    }
  };

  const releaseBot = async () => {
    if (!id) return;
    await api.release(id);
    loadChats();
  };

  return (
      <div className="h-full flex bg-[#F7F8FA] dark:bg-slate-900">

        {/* LEFT SIDEBAR: CHAT LIST */}
        <div className="w-80 bg-white dark:bg-slate-950 border-r border-slate-200 dark:border-slate-800 flex flex-col shrink-0">
          <div className="p-4 border-b border-slate-100 dark:border-slate-800">
            <h2 className="text-sm font-bold text-slate-900 dark:text-slate-100 uppercase tracking-wide mb-1">
              {t(i18nKeys.chats.title)}
            </h2>
            <p className="text-xs text-slate-500 dark:text-slate-400 font-medium">
              {t(i18nKeys.chats.count)} {chats.length}
            </p>
          </div>
          <div className="flex-1 overflow-y-auto">
            {chats.length === 0 && (
                <div className="p-4 text-xs text-slate-400 text-center">
                  {t(i18nKeys.chats.empty)}
                </div>
            )}
            {chats.map((row) => {
              const active = row.conversation.id === id;
              return (
                  <button
                      key={row.conversation.id}
                      onClick={() => nav(`/chats/${row.conversation.id}`)}
                      className={`w-full text-left p-3 border-b border-slate-50 dark:border-slate-800 transition-colors flex items-start gap-3 ${
                          active
                              ? "bg-blue-50/50 dark:bg-blue-600/20 border-l-4 border-l-blue-600 pl-[11px]"
                              : "hover:bg-slate-50 dark:hover:bg-slate-900 border-l-4 border-l-transparent pl-[15px]"
                      }`}
                  >
                    <div className="w-9 h-9 rounded-full bg-slate-100 dark:bg-slate-800 grid place-items-center text-slate-600 dark:text-slate-300 font-medium text-sm shrink-0">
                      {(row.patient?.name || row.patient?.phone || "?").charAt(0).toUpperCase()}
                    </div>
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center justify-between gap-2 mb-0.5">
                        <div className="font-medium text-sm text-slate-900 dark:text-slate-100 truncate">
                          {row.patient?.name || row.patient?.phone || t(i18nKeys.chats.no_patient)}
                        </div>
                        {row.conversation.status === "handoff" && (
                            <span className="text-[10px] px-1.5 py-0.5 rounded-full bg-amber-50 dark:bg-amber-900/30 text-amber-700 dark:text-amber-400 border border-amber-100 dark:border-amber-800 font-bold uppercase tracking-wide shrink-0">
                              {t(i18nKeys.chats.operator)}
                            </span>
                        )}
                      </div>
                      <div className="text-xs text-slate-500 dark:text-slate-400 truncate">
                        {row.last_message?.body || "—"}
                      </div>
                    </div>
                  </button>
              );
            })}
          </div>
        </div>

        {/* RIGHT PANEL: CONVERSATION */}
        <div className="flex-1 flex flex-col bg-[#F7F8FA] dark:bg-slate-900">
          {!id ? (
              <div className="h-full flex flex-col items-center justify-center text-slate-400">
                <div className="w-16 h-16 bg-slate-100 dark:bg-slate-800 rounded-full grid place-items-center text-2xl mb-4">
                  💬
                </div>
                <p className="text-sm font-medium">
                  {t(i18nKeys.chats.select_chat)}
                </p>
              </div>
          ) : (
              <>
                {/* Header */}
                <div className="h-16 bg-white dark:bg-slate-950 border-b border-slate-200 dark:border-slate-800 flex items-center justify-between px-6 shrink-0">
                  <div>
                    <div className="font-bold text-slate-900 dark:text-slate-100">
                      {current?.patient?.name || current?.patient?.phone || t("chats.chat_fallback")}
                    </div>
                    <div className="text-xs text-slate-500 dark:text-slate-400 font-medium">
                      {current?.patient?.phone}
                      {current?.conversation.status === "handoff" && (
                          <span className="text-amber-600 dark:text-amber-400 ml-2">· {t(i18nKeys.chats.operator)}</span>
                      )}
                    </div>
                  </div>
                  {current?.conversation.status === "handoff" && (
                      <button
                          onClick={releaseBot}
                          className="px-3 py-1.5 text-xs font-medium text-slate-600 dark:text-slate-300 bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors"
                      >
                        {t(i18nKeys.chats.return_to_bot)}
                      </button>
                  )}
                </div>

                {/* Messages Area */}
                <div ref={scrollRef} className="flex-1 overflow-y-auto px-6 py-4 space-y-4">
                  {loading && (
                      <div className="text-center text-xs text-slate-400">
                        {t(i18nKeys.chats.loading)}
                      </div>
                  )}
                  {messages.map((m) => {
                    const isOut = m.direction === "outbound";
                    const byOperator = m.sender === "operator";
                    return (
                        <div
                            key={m.id}
                            className={`flex ${isOut ? "justify-end" : "justify-start"}`}
                        >
                          <div
                              className={`max-w-[75%] rounded-2xl px-4 py-2.5 text-sm shadow-sm ${
                                  isOut
                                      ? byOperator
                                          ? "bg-slate-800 dark:bg-slate-600 text-white"
                                          : "bg-blue-600 text-white"
                                      : "bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 text-slate-800 dark:text-slate-100"
                              }`}
                          >
                            <div className="whitespace-pre-wrap leading-relaxed">
                              {m.body}
                            </div>
                            <div className={`text-[10px] mt-1.5 font-medium ${
                                isOut ? "text-white/70" : "text-slate-400"
                            }`}>
                              {isOut
                                  ? (byOperator ? t(i18nKeys.sender.operator) : t(i18nKeys.sender.aigul))
                                  : t(i18nKeys.sender.patient)}
                              {" · "}
                              {new Date(m.created_at).toLocaleString("ru-RU", {
                                hour: "2-digit",
                                minute: "2-digit",
                                day: "2-digit",
                                month: "2-digit",
                              })}
                            </div>
                          </div>
                        </div>
                    );
                  })}
                </div>

                {/* Input Area */}
                <div className="bg-white dark:bg-slate-950 border-t border-slate-200 dark:border-slate-800 p-4">
                  <div className="flex gap-3">
                    <input
                        className="flex-1 px-4 py-2.5 bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-slate-700 rounded-lg text-sm text-slate-900 dark:text-slate-100 placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 transition-all"
                        placeholder={t(i18nKeys.chats.placeholder)}
                        value={input}
                        onChange={(e) => setInput(e.target.value)}
                        onKeyDown={(e) => {
                          if (e.key === "Enter" && !e.shiftKey) {
                            e.preventDefault();
                            send();
                          }
                        }}
                    />
                    <button
                        onClick={send}
                        className="px-6 py-2.5 bg-blue-600 text-white text-sm font-medium rounded-lg hover:bg-blue-700 transition-colors shadow-sm disabled:opacity-70 disabled:cursor-not-allowed"
                        disabled={sending || !input.trim()}
                    >
                      {sending ? t(i18nKeys.chats.sending) : t(i18nKeys.chats.send)}
                    </button>
                  </div>
                  <p className="text-[11px] text-slate-400 mt-2">
                    {t(i18nKeys.chats.info_text)}
                  </p>
                </div>
              </>
          )}
        </div>
      </div>
  );
}
