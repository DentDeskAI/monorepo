import { useEffect, useMemo, useRef, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { api } from "../api/client";

export default function Chats() {
  const { id } = useParams();
  const nav = useNavigate();
  const [chats, setChats] = useState([]);
  const [messages, setMessages] = useState([]);
  const [loading, setLoading] = useState(false);
  const [sending, setSending] = useState(false);
  const [input, setInput] = useState("");
  const scrollRef = useRef(null);

  // Загрузка списка чатов
  const loadChats = () => api.chats().then(setChats).catch(() => {});
  useEffect(() => { loadChats(); }, []);

  // Загрузка сообщений при смене чата
  useEffect(() => {
    if (!id) { setMessages([]); return; }
    setLoading(true);
    api.messages(id).then((msgs) => setMessages(msgs || [])).finally(() => setLoading(false));
  }, [id]);

  // SSE — обновляем на лету
  useEffect(() => {
    const unsub = api.subscribe((type, payload) => {
      if (type === "message") {
        // обновить список чатов (последнее сообщение + сортировка)
        loadChats();
        // если сообщение в текущем открытом чате — добавить
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

  // автоскролл вниз
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
      alert("Не удалось отправить: " + e.message);
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
    <div className="h-full flex">
      {/* Список чатов */}
      <div className="w-80 shrink-0 border-r border-slate-200 bg-white flex flex-col">
        <div className="p-4 border-b border-slate-200">
          <h2 className="font-semibold text-slate-900">Чаты</h2>
          <p className="text-xs text-slate-500">{chats.length} диалогов</p>
        </div>
        <div className="flex-1 overflow-auto">
          {chats.length === 0 && (
            <div className="p-4 text-sm text-slate-500">Пока нет чатов</div>
          )}
          {chats.map((row) => {
            const active = row.conversation.id === id;
            return (
              <button
                key={row.conversation.id}
                onClick={() => nav(`/chats/${row.conversation.id}`)}
                className={`w-full text-left p-3 border-b border-slate-100 transition ${
                  active ? "bg-brand-50" : "hover:bg-slate-50"
                }`}
              >
                <div className="flex items-center gap-2">
                  <div className="w-9 h-9 rounded-full bg-slate-100 grid place-items-center text-slate-600 font-medium shrink-0">
                    {(row.patient?.name || row.patient?.phone || "?").charAt(0).toUpperCase()}
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center justify-between gap-2">
                      <div className="font-medium text-sm text-slate-900 truncate">
                        {row.patient?.name || row.patient?.phone}
                      </div>
                      {row.conversation.status === "handoff" && (
                        <span className="text-[10px] px-1.5 py-0.5 rounded bg-amber-100 text-amber-700 shrink-0">
                          оператор
                        </span>
                      )}
                    </div>
                    <div className="text-xs text-slate-500 truncate">
                      {row.last_message?.body || "—"}
                    </div>
                  </div>
                </div>
              </button>
            );
          })}
        </div>
      </div>

      {/* Диалог */}
      <div className="flex-1 flex flex-col bg-slate-50">
        {!id ? (
          <div className="h-full grid place-items-center text-slate-400">
            Выберите чат слева
          </div>
        ) : (
          <>
            <div className="px-6 py-3 bg-white border-b border-slate-200 flex items-center justify-between">
              <div>
                <div className="font-medium text-slate-900">
                  {current?.patient?.name || current?.patient?.phone || "Чат"}
                </div>
                <div className="text-xs text-slate-500">
                  {current?.patient?.phone}
                  {current?.conversation.status === "handoff" && " · режим оператора"}
                </div>
              </div>
              {current?.conversation.status === "handoff" && (
                <button onClick={releaseBot} className="btn btn-secondary text-xs">
                  Вернуть боту
                </button>
              )}
            </div>

            <div ref={scrollRef} className="flex-1 overflow-auto px-6 py-4 space-y-2">
              {loading && <div className="text-sm text-slate-400">Загружаем...</div>}
              {messages.map((m) => {
                const isOut = m.direction === "outbound";
                const byOperator = m.sender === "operator";
                return (
                  <div
                    key={m.id}
                    className={`flex ${isOut ? "justify-end" : "justify-start"}`}
                  >
                    <div
                      className={`max-w-[70%] rounded-2xl px-4 py-2 text-sm ${
                        isOut
                          ? byOperator
                            ? "bg-slate-800 text-white"
                            : "bg-brand-600 text-white"
                          : "bg-white border border-slate-200 text-slate-800"
                      }`}
                    >
                      <div className="whitespace-pre-wrap">{m.body}</div>
                      <div className={`text-[10px] mt-1 ${isOut ? "text-white/70" : "text-slate-400"}`}>
                        {isOut ? (byOperator ? "оператор" : "Айгуль") : "пациент"}
                        {" · "}
                        {new Date(m.created_at).toLocaleString("ru-RU", {
                          hour: "2-digit", minute: "2-digit",
                          day: "2-digit", month: "2-digit",
                        })}
                      </div>
                    </div>
                  </div>
                );
              })}
            </div>

            <div className="p-4 bg-white border-t border-slate-200">
              <div className="flex gap-2">
                <input
                  className="input"
                  placeholder="Написать как оператор (бот замолчит)..."
                  value={input}
                  onChange={(e) => setInput(e.target.value)}
                  onKeyDown={(e) => {
                    if (e.key === "Enter" && !e.shiftKey) {
                      e.preventDefault();
                      send();
                    }
                  }}
                />
                <button onClick={send} className="btn btn-primary" disabled={sending || !input.trim()}>
                  Отправить
                </button>
              </div>
              <p className="text-[11px] text-slate-400 mt-1">
                Отправка сообщения переключает чат в режим оператора. Чтобы вернуть бота — кнопка сверху.
              </p>
            </div>
          </>
        )}
      </div>
    </div>
  );
}
