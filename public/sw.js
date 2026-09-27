const CACHE_NAME = "projecao-static-v3";
const STATIC_ASSETS = ["/manifest.webmanifest", "/icon.svg"];

self.addEventListener("install", (event) => {
  event.waitUntil(caches.open(CACHE_NAME).then((cache) => cache.addAll(STATIC_ASSETS)));
  self.skipWaiting();
});

self.addEventListener("activate", (event) => {
  event.waitUntil(
    caches.keys().then((keys) =>
      Promise.all(keys.filter((key) => key !== CACHE_NAME).map((key) => caches.delete(key))),
    ),
  );
  self.clients.claim();
});

self.addEventListener("fetch", (event) => {
  const url = new URL(event.request.url);
  if (event.request.method !== "GET" || url.origin !== self.location.origin) return;

  // Authenticated HTML and API responses must never be stored in a shared cache.
  if (event.request.mode === "navigate" || url.pathname.startsWith("/api/")) return;

  if (STATIC_ASSETS.includes(url.pathname) || url.pathname.startsWith("/_next/static/")) {
    event.respondWith(
      caches.match(event.request).then(async (cached) => {
        if (cached) return cached;

        const response = await fetch(event.request);
        if (response.ok) {
          // Clone before the browser starts consuming the response body.
          const cacheCopy = response.clone();
          event.waitUntil(
            caches.open(CACHE_NAME)
              .then((cache) => cache.put(event.request, cacheCopy))
              .catch(() => undefined),
          );
        }
        return response;
      }),
    );
  }
});
