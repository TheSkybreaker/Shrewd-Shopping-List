import { cleanupOutdatedCaches, precacheAndRoute } from 'workbox-precaching';

declare let self: ServiceWorkerGlobalScope;

// Only the app shell is precached. /api requests are never in the manifest, so they always reach
// the network. Query parameters such as ?id= are ignored: every list shares the same page.
precacheAndRoute(self.__WB_MANIFEST, { ignoreURLParametersMatching: [/.*/] });
cleanupOutdatedCaches();

// "Aggiorna" in the new version toast activates the waiting service worker.
self.addEventListener('message', (event) => {
  if (event.data?.type === 'SKIP_WAITING') void self.skipWaiting();
});
