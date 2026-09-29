"use client";

import { useEffect, useState } from "react";
import { CheckCircle2, CircleAlert, X } from "lucide-react";
import { API_NOTICE_EVENT, type ApiNotice } from "@/lib/api/notice-event";

type Notice = ApiNotice & { id: number };
let nextId = 0;

export function ApiNotifications() {
  const [notices, setNotices] = useState<Notice[]>([]);

  useEffect(() => {
    const onNotice = (event: Event) => {
      const detail = (event as CustomEvent<ApiNotice>).detail;
      if (!detail?.message) return;
      const id = ++nextId;
      setNotices((current) => [...current.slice(-2), { ...detail, id }]);
    };
    window.addEventListener(API_NOTICE_EVENT, onNotice);
    return () => window.removeEventListener(API_NOTICE_EVENT, onNotice);
  }, []);

  useEffect(() => {
    if (!notices.length) return;
    const timers = notices.map(({ id, type }) => window.setTimeout(() => {
      setNotices((current) => current.filter((notice) => notice.id !== id));
    }, type === "error" ? 8000 : 5000));
    return () => timers.forEach(window.clearTimeout);
  }, [notices]);

  return <div className="api-notifications" aria-label="Notificações">
    {notices.map((notice) => <div key={notice.id} className={`api-notice ${notice.type}`} role={notice.type === "error" ? "alert" : "status"}>
      {notice.type === "success" ? <CheckCircle2 size={20} aria-hidden="true" /> : <CircleAlert size={20} aria-hidden="true" />}
      <span>{notice.message}</span>
      <button type="button" aria-label="Dispensar notificação" onClick={() => setNotices((current) => current.filter((item) => item.id !== notice.id))}><X size={18} /></button>
    </div>)}
  </div>;
}
