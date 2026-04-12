const CACHE = 'physicscopilot-v1';
const APP_SHELL = ['/', '/index.html', '/manifest.json', '/favicon.svg'];
const OFFLINE_PAGE = '/offline.html';

self.addEventListener('install', event => {
  event.waitUntil(
    caches.open(CACHE).then(cache =>
      cache.addAll([...APP_SHELL, OFFLINE_PAGE])
    ).then(() => self.skipWaiting())
  );
});

self.addEventListener('activate', event => {
  event.waitUntil(
    caches.keys().then(keys =>
      Promise.all(keys.filter(k => k !== CACHE).map(k => caches.delete(k)))
    ).then(() => self.clients.claim())
  );
});

self.addEventListener('fetch', event => {
  if (event.request.method !== 'GET') return;

  event.respondWith(
    caches.match(event.request).then(cached => {
      if (cached) return cached;

      return fetch(event.request).then(response => {
        if (!response || response.status !== 200 || response.type !== 'basic') {
          return response;
        }
        const clone = response.clone();
        caches.open(CACHE).then(cache => cache.put(event.request, clone));
        return response;
      }).catch(() =>
        caches.match(OFFLINE_PAGE)
      );
    })
  );
});
