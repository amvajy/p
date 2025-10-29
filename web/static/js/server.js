(function(){
  const API_BASE = 'http://localhost:8080/api';
  let authToken = localStorage.getItem('pxe_auth_token') || '';
  const els = {
    tokenInput: document.getElementById('authToken'),
    saveTokenBtn: document.getElementById('saveTokenBtn'),
    tokenStatus: document.getElementById('tokenStatus'),
    title: document.getElementById('title'),
    detail: document.getElementById('detail')
  };
  els.tokenInput.value = authToken;
  els.saveTokenBtn.addEventListener('click', () => {
    authToken = els.tokenInput.value.trim();
    localStorage.setItem('pxe_auth_token', authToken);
    els.tokenStatus.textContent = authToken ? '已保存' : '未设置';
  });

  function headers() {
    const h = { 'Content-Type': 'application/json' };
    if (authToken) h['Authorization'] = 'Bearer ' + authToken;
    return h;
  }

  function getSerial() {
    const sp = new URLSearchParams(location.search);
    return sp.get('serial') || '';
  }

  async function load() {
    const serial = getSerial();
    if (!serial) {
      els.title.textContent = '未提供序列号 serial';
      els.detail.textContent = '请在 URL 查询参数中提供 serial，例如 server.html?serial=ABC123';
      return;
    }
    els.title.textContent = '服务器 ' + serial;
    els.detail.textContent = '加载中...';
    try {
      const res = await fetch(API_BASE + '/servers/' + encodeURIComponent(serial), { headers: headers() });
      if (!res.ok) throw new Error('HTTP ' + res.status);
      const obj = await res.json();
      els.detail.textContent = JSON.stringify(obj, null, 2);
    } catch (e) {
      els.detail.textContent = '加载失败：' + e.message;
    }
  }

  // 操作按钮
  document.getElementById('btnConfirm').addEventListener('click', async () => {
    const serial = getSerial();
    try {
      const res = await fetch(API_BASE + '/servers/' + encodeURIComponent(serial) + '/confirm', { method: 'POST', headers: headers() });
      if (!res.ok) throw new Error('HTTP ' + res.status);
      alert('已确认');
      load();
    } catch (e) {
      alert('失败：' + e.message);
    }
  });
  document.getElementById('btnInstall').addEventListener('click', async () => {
    const serial = getSerial();
    try {
      const res = await fetch(API_BASE + '/servers/' + encodeURIComponent(serial) + '/install', { method: 'POST', headers: headers() });
      if (!res.ok) throw new Error('HTTP ' + res.status);
      alert('已标记安装');
      load();
    } catch (e) {
      alert('失败：' + e.message);
    }
  });
  document.getElementById('btnApplyCfg').addEventListener('click', async () => {
    const serial = getSerial();
    const cfgId = (document.getElementById('applyCfgIdDetail').value||'').trim();
    if (!cfgId) { alert('请填写模板ID'); return; }
    try {
      const res = await fetch(API_BASE + '/configs/' + encodeURIComponent(cfgId) + '/apply?serial=' + encodeURIComponent(serial), { method: 'POST', headers: headers() });
      if (!res.ok) throw new Error('HTTP ' + res.status);
      const j = await res.json();
      alert(j.message||'已应用');
      load();
    } catch (e) {
      alert('失败：' + e.message);
    }
  });

  load();
})();
