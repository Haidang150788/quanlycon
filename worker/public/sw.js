// Service worker: nhan Web Push tu server, hien thong bao, bam vao thi mo trang dieu khien.
self.addEventListener('install', () => self.skipWaiting());
self.addEventListener('activate', (e) => e.waitUntil(self.clients.claim()));
self.addEventListener('push', (e) => {
  let tin = {};
  try { tin = e.data ? e.data.json() : {}; } catch { tin = { noi_dung: e.data && e.data.text() }; }
  e.waitUntil(self.registration.showNotification(tin.tieu_de || 'Máy của con', {
    body: tin.noi_dung || 'Có yêu cầu mới, mở để xem',
    tag: tin.the || 'qlc',
    renotify: true,
    requireInteraction: true,
    icon: '/icon-192.png',
    badge: '/icon-192.png',
    data: { url: '/' },
  }));
});
self.addEventListener('notificationclick', (e) => {
  e.notification.close();
  e.waitUntil((async () => {
    const ds = await self.clients.matchAll({ type: 'window', includeUncontrolled: true });
    for (const c of ds) { if ('focus' in c) { c.navigate('/'); return c.focus(); } }
    return self.clients.openWindow('/');
  })());
});
