// Danh muc nen tang: ten mien chan tren trinh duyet + ten tien trinh (.exe) bi dong.
// Nam o SERVER de them/sua nen tang ma khong phai cai lai phan mem tren may con.
// Ten mien: Chrome/Edge URLBlocklist chan ca ten mien con (facebook.com => m.facebook.com),
// con file hosts thi khong co wildcard nen liet ke them cac ten mien con hay dung.
// Tien trinh: so khop KHONG phan biet hoa thuong, chi ten file.

export const NHOM = [
  { ma: 'mxh', ten: 'Mạng xã hội' },
  { ma: 'game', ten: 'Game' },
];

export const DANH_MUC = [
  { ma: 'facebook', ten: 'Facebook', nhom: 'mxh', mau: '#1877f2', chu: 'f',
    mien: ['facebook.com', 'www.facebook.com', 'm.facebook.com', 'web.facebook.com', 'fb.com', 'fb.watch', 'messenger.com', 'www.messenger.com'],
    exe: ['Messenger.exe'] },
  { ma: 'tiktok', ten: 'TikTok', nhom: 'mxh', mau: '#111', chu: 'Tt',
    mien: ['tiktok.com', 'www.tiktok.com', 'm.tiktok.com', 'vm.tiktok.com', 'vt.tiktok.com', 'tiktokv.com'],
    exe: ['TikTok.exe'] },
  { ma: 'youtube', ten: 'YouTube', nhom: 'mxh', mau: '#ff0000', chu: '▶',
    // Them: trang xem YouTube "thay the" (extension khong kiem duoc kenh) => luon chan cung YouTube.
    mien: ['youtube.com', 'www.youtube.com', 'm.youtube.com', 'music.youtube.com', 'youtu.be', 'youtube-nocookie.com', 'www.youtube-nocookie.com',
      'yewtu.be', 'invidious.io', 'inv.nadeko.net', 'invidious.nerdvpn.de', 'invidious.private.coffee', 'invidious.f5.si', 'iv.melmac.space',
      'piped.video', 'piped.kavin.rocks', 'piped.private.coffee', 'poketube.fun', 'viewtube.io'],
    // Co danh sach kenh duoc xem + extension dang chay trong trinh duyet => may con BO cac ten mien nay khoi
    // danh sach chan cua trinh duyet do (extension tu chan theo kenh). Phan con lai cua `mien` (YouTube Music,
    // trang thay the) van chan.
    mien_theo_kenh: ['youtube.com', 'www.youtube.com', 'm.youtube.com', 'youtu.be', 'youtube-nocookie.com', 'www.youtube-nocookie.com'],
    exe: ['FreeTube.exe', 'Grayjay.exe'] },
  { ma: 'instagram', ten: 'Instagram', nhom: 'mxh', mau: '#e1306c', chu: 'Ig',
    mien: ['instagram.com', 'www.instagram.com', 'm.instagram.com'], exe: ['Instagram.exe'] },
  { ma: 'threads', ten: 'Threads', nhom: 'mxh', mau: '#111', chu: '@',
    mien: ['threads.net', 'www.threads.net', 'threads.com', 'www.threads.com'], exe: [] },
  { ma: 'x', ten: 'X (Twitter)', nhom: 'mxh', mau: '#111', chu: 'X',
    mien: ['x.com', 'www.x.com', 'twitter.com', 'www.twitter.com', 'mobile.twitter.com'], exe: [] },
  { ma: 'telegram', ten: 'Telegram', nhom: 'mxh', mau: '#27a7e7', chu: 'Tg',
    mien: ['telegram.org', 'web.telegram.org', 't.me'], exe: ['Telegram.exe'] },
  { ma: 'discord', ten: 'Discord', nhom: 'mxh', mau: '#5865f2', chu: 'Dc',
    mien: ['discord.com', 'www.discord.com', 'discord.gg', 'discordapp.com', 'www.discordapp.com'],
    exe: ['Discord.exe', 'DiscordPTB.exe', 'DiscordCanary.exe'] },

  { ma: 'roblox', ten: 'Roblox', nhom: 'game', mau: '#393b3d', chu: 'Rb',
    mien: ['roblox.com', 'www.roblox.com', 'web.roblox.com'],
    exe: ['RobloxPlayerBeta.exe', 'RobloxPlayerLauncher.exe', 'RobloxStudioBeta.exe', 'RobloxPlayerInstaller.exe', 'Windows10Universal.exe'] },
  { ma: 'minecraft', ten: 'Minecraft', nhom: 'game', mau: '#5b8731', chu: 'Mc',
    mien: [], exe: ['Minecraft.exe', 'MinecraftLauncher.exe', 'Minecraft.Windows.exe', 'TLauncher.exe'] },
  { ma: 'garena', ten: 'Garena', nhom: 'game', mau: '#ee4d2d', chu: 'G',
    mien: ['garena.com', 'garena.vn', 'www.garena.vn'], exe: ['Garena.exe', 'GarenaMessenger.exe', 'gxxcef.exe'] },
  { ma: 'riot', ten: 'Riot (LMHT, Valorant)', nhom: 'game', mau: '#d13639', chu: 'R',
    mien: ['riotgames.com', 'www.riotgames.com', 'leagueoflegends.com', 'www.leagueoflegends.com', 'playvalorant.com'],
    exe: ['RiotClientServices.exe', 'RiotClientUx.exe', 'Riot Client.exe', 'LeagueClient.exe', 'LeagueClientUx.exe', 'League of Legends.exe', 'VALORANT.exe', 'VALORANT-Win64-Shipping.exe'] },
  { ma: 'steam', ten: 'Steam / Epic', nhom: 'game', mau: '#1b2838', chu: 'St',
    mien: ['store.steampowered.com', 'steamcommunity.com', 'epicgames.com', 'store.epicgames.com'],
    exe: ['steam.exe', 'steamwebhelper.exe', 'EpicGamesLauncher.exe'] },
  { ma: 'gameweb', ten: 'Game trên web', nhom: 'game', mau: '#7b4fd6', chu: 'Gw',
    mien: ['poki.com', 'www.poki.com', 'crazygames.com', 'www.crazygames.com', 'y8.com', 'www.y8.com', 'friv.com', 'www.friv.com', 'gamevui.vn', 'game24h.vn', 'miniclip.com', 'coolmathgames.com', 'now.gg', 'krunker.io', 'agar.io', 'slither.io'],
    exe: [] },
];

export const MA_HOP_LE = new Set(DANH_MUC.map((n) => n.ma));
