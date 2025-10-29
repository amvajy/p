(function(){
  const API_BASE = 'http://localhost:8080/api';
  let authToken = localStorage.getItem('pxe_auth_token') || '';

  const els = {
    tokenInput: document.getElementById('authToken'),
    saveTokenBtn: document.getElementById('saveTokenBtn'),
    tokenStatus: document.getElementById('tokenStatus'),
    offset: document.getElementById('offset'),
    limit: document.getElementById('limit'),
    order: document.getElementById('order'),
    loadBtn: document.getElementById('loadBtn'),
    logs: document.getElementById('logs')
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

  let lastLogs = [];
  function applyFilters() {
    const m = (document.getElementById('fMethod').value||'').trim().toUpperCase();
    const p = (document.getElementById('fPath').value||'').trim();
    const a = (document.getElementById('fAction').value||'').trim();
    const s = (document.getElementById('fStatus').value||'').trim();
    let arr = lastLogs.slice();
    if (m) arr = arr.filter(x => String(x.method||'').toUpperCase()===m);
    if (p) arr = arr.filter(x => String(x.path||'').includes(p));
    if (a) arr = arr.filter(x => String(x.action||'').includes(a));
    if (s) arr = arr.filter(x => String(x.status||'').includes(s));
    els.logs.textContent = JSON.stringify(arr, null, 2);
    return arr;
  }

  async function load() {
    els.logs.textContent = '加载中...';
    const qs = new URLSearchParams({
      offset: String(els.offset.value||0),
      limit: String(els.limit.value||100),
      order: String(els.order.value||'desc')
    });
    try {
      const res = await fetch(API_BASE + '/audit/logs?' + qs.toString(), { headers: headers() });
      if (!res.ok) throw new Error('HTTP ' + res.status);
      lastLogs = await res.json();
      applyFilters();
    } catch (e) {
      els.logs.textContent = '加载失败：' + e.message;
    }
  }

  els.loadBtn.addEventListener('click', load);
  ['fMethod','fPath','fAction','fStatus'].forEach(id => {
    const el = document.getElementById(id);
    if (el) el.addEventListener('input', applyFilters);
  });
  document.getElementById('exportBtn').addEventListener('click', () => {
    const arr = applyFilters();
    const blob = new Blob([JSON.stringify(arr, null, 2)], {type: 'application/json'});
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = 'audit_logs.json';
    a.click();
    URL.revokeObjectURL(url);
  });

  load();
})();
