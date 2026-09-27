"use client";

import { useEffect, useState } from "react";

export function PwaRegister() {
  const [needsReload, setNeedsReload] = useState(false);

  useEffect(() => {
    if (!("serviceWorker" in navigator)) return;

    if (process.env.NODE_ENV === "development") {
      // A previous production registration can cache stale Next.js assets in dev.
      navigator.serviceWorker.getRegistrations()
        .then((registrations) => Promise.all(registrations.map((registration) => registration.unregister())))
        .then(() => {
          // Unregistering does not release the worker controlling this tab until a reload.
          if (navigator.serviceWorker.controller) setNeedsReload(true);
        })
        .catch(() => undefined);
      return;
    }

    navigator.serviceWorker.register("/sw.js").catch(() => undefined);
  }, []);

  if (!needsReload) return null;

  return <div className="dev-cache-notice" role="status">
    <span>Esta aba ainda usa uma versão antiga do app.</span>
    <button type="button" onClick={() => window.location.reload()}>Recarregar versão atual</button>
  </div>;
}
