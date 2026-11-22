// Chay trong "the gioi" cua trang YouTube (world: MAIN) de doc trinh phat: video dang mo thuoc kenh nao.
// Gui sang noi-dung.js bang postMessage. Khong sua gi tren trang.
(function () {
  function doc() {
    let pr = null;
    try {
      const p = document.getElementById('movie_player');
      if (p && typeof p.getPlayerResponse === 'function') pr = p.getPlayerResponse();
    } catch { /* trinh phat chua san sang */ }
    if (!pr || !pr.videoDetails) pr = window.ytInitialPlayerResponse;
    const vd = pr && pr.videoDetails;
    if (!vd || !vd.videoId) return;
    window.postMessage({ __qlc_trinh_phat: { id: vd.videoId, kenh_id: vd.channelId || '', tac_gia: vd.author || '' } }, location.origin);
  }
  setInterval(doc, 400);
  document.addEventListener('yt-navigate-finish', doc);
})();
