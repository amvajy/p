(function(){
  const API_BASE = 'http://localhost:8080/api';
  let authToken = localStorage.getItem('pxe_auth_token') || '';

  const els = {
    tokenInput: document.getElementById('authToken'),
    saveTokenBtn: document.getElementById('saveTokenBtn'),
    tokenStatus: document.getElementById('tokenStatus'),

    refreshServersBtn: document.getElementById('refreshServersBtn'),
    serversTableBody: document.querySelector('#serversTable tbody'),

    refreshConfigsBtn: document.getElementById('refreshConfigsBtn'),
    configsTableBody: document.querySelector('#configsTable tbody'),

    configForm: document.getElementById('configForm'),
    cfgId: document.getElementById('cfgId'),
    cfgName: document.getElementById('cfgName'),
    cfgDescription: document.getElementById('cfgDescription'),
    cfgSystemType: document.getElementById('cfgSystemType'),
    cfgSystemVersion: document.getElementById('cfgSystemVersion'),
    cfgContent: document.getElementById('cfgContent'),
    cfgKernelParams: document.getElementById('cfgKernelParams'),
    cfgPackages: document.getElementById('cfgPackages'),

    applyForm: document.getElementById('applyForm'),
    applyCfgId: document.getElementById('applyCfgId'),
    applySerial: document.getElementById('applySerial'),

    loadAllBtn: document.getElementById('loadAllBtn'),
    debugOutput: document.getElementById('debugOutput')
  };

  els.tokenInput.value = authToken;
  els.saveTokenBtn.addEventListener('click', () => {
    authToken = els.tokenInput.value.trim();
    localStorage.setItem('pxe_auth_token', authToken);
    els.tokenStatus.textContent = authToken ? '已保存' : '未设置';
    pingBackend();
  });

  function headers() {
    const h = { 'Content-Type': 'application/json' };
    if (authToken) {
      h['Authorization'] = 'Bearer ' + authToken;
    }
    return h;
  }

  async function apiGet(path) {
    const res = await fetch(API_BASE + path, { headers: headers() });
    if (!res.ok) throw new Error('HTTP ' + res.status);
    return res.json();
  }
  async function apiPost(path, data) {
    const res = await fetch(API_BASE + path, { method: 'POST', headers: headers(), body: JSON.stringify(data||{}) });
    if (!res.ok) throw new Error('HTTP ' + res.status);
    return res.json();
  }
  async function apiPut(path, data) {
    const res = await fetch(API_BASE + path, { method: 'PUT', headers: headers(), body: JSON.stringify(data||{}) });
    if (!res.ok) throw new Error('HTTP ' + res.status);
    return res.text();
  }

  // servers
  async function loadServers(status='pending') {
    try {
      const data = await apiGet('/servers?status=' + encodeURIComponent(status));
      els.serversTableBody.innerHTML = '';
      data.forEach(row => {
        const tr = document.createElement('tr');
        tr.innerHTML = `
          <td>${row.serial||''}</td>
          <td>${row.hostname||''}</td>
          <td>${row.ipAddress||''}</td>
          <td>${row.macAddress||''}</td>
          <td>${row.status||''}</td>
          <td>
            <button data-serial="${row.serial}" class="btn-confirm">确认</button>
            <button data-serial="${row.serial}" class="btn-install">标记安装</button>
            <button data-serial="${row.serial}" class="btn-detail">详情</button>
          </td>
        `;
        els.serversTableBody.appendChild(tr);
      });
    } catch (e) {
      els.serversTableBody.innerHTML = `<tr><td colspan="6" class="error">加载失败：${e.message}</td></tr>`;
    }
  }

  els.refreshServersBtn.addEventListener('click', () => {
    const statusSel = document.getElementById('serversStatus');
    const status = statusSel ? statusSel.value : 'pending';
    loadServers(status);
  });
  els.serversTableBody.addEventListener('click', async (ev) => {
    const btn = ev.target.closest('button');
    if (!btn) return;
    const serial = btn.getAttribute('data-serial');
    try {
      if (btn.classList.contains('btn-confirm')) {
        const j = await apiPost(`/servers/${encodeURIComponent(serial)}/confirm`);
        alert(j.message||'已确认');
      } else if (btn.classList.contains('btn-install')) {
        const j = await apiPost(`/servers/${encodeURIComponent(serial)}/install`);
        alert(j.message||'已标记安装');
      } else if (btn.classList.contains('btn-detail')) {
        if (serial) {
          window.location.href = '/templates/server.html?serial=' + encodeURIComponent(serial);
        }
        return;
      }
      await loadServers('pending');
    } catch (e) {
      alert('操作失败：' + e.message);
    }
  });

  // configs
  async function loadConfigs() {
    try {
      const data = await apiGet('/configs');
      els.configsTableBody.innerHTML = '';
      data.forEach(row => {
        const tr = document.createElement('tr');
        tr.innerHTML = `
          <td>${row.id||''}</td>
          <td>${row.name||''}</td>
          <td>${row.systemType||''}</td>
          <td>${row.systemVersion||''}</td>
          <td>
            <button data-id="${row.id}" class="btn-fill">填入</button>
          </td>
        `;
        els.configsTableBody.appendChild(tr);
      });
    } catch (e) {
      els.configsTableBody.innerHTML = `<tr><td colspan="5" class="error">加载失败：${e.message}</td></tr>`;
    }
  }

  els.refreshConfigsBtn.addEventListener('click', loadConfigs);
  els.configsTableBody.addEventListener('click', async (ev) => {
    const btn = ev.target.closest('button');
    if (!btn) return;
    if (btn.classList.contains('btn-fill')) {
      const id = btn.getAttribute('data-id');
      try {
        const cfg = await apiGet('/configs/' + encodeURIComponent(id));
        els.cfgId.value = cfg.id || '';
        els.cfgName.value = cfg.name || '';
        els.cfgDescription.value = cfg.description || '';
        els.cfgSystemType.value = cfg.systemType || 'CentOS';
        els.cfgSystemVersion.value = cfg.systemVersion || '';
        els.cfgContent.value = cfg.configContent || '';
        els.cfgKernelParams.value = cfg.kernelParams || '';
        els.cfgPackages.value = cfg.packages || '';
      } catch (e) {
        alert('获取配置失败：' + e.message);
      }
    }
  });

  els.configForm.addEventListener('submit', async (ev) => {
    ev.preventDefault();
    const id = (els.cfgId.value||'').trim();
    const payload = {
      name: (els.cfgName.value||'').trim(),
      description: (els.cfgDescription.value||'').trim(),
      systemType: els.cfgSystemType.value,
      systemVersion: (els.cfgSystemVersion.value||'').trim(),
      configContent: els.cfgContent.value,
      kernelParams: els.cfgKernelParams.value,
      packages: els.cfgPackages.value
    };
    try {
      if (id) {
        await apiPut('/configs/' + encodeURIComponent(id), payload);
        alert('已更新');
      } else {
        const j = await apiPost('/configs', payload);
        alert('已创建：ID ' + j.id);
      }
      await loadConfigs();
    } catch (e) {
      alert('提交失败：' + e.message);
    }
  });

  els.applyForm.addEventListener('submit', async (ev) => {
    ev.preventDefault();
    const cfgId = (els.applyCfgId.value||'').trim();
    const serial = (els.applySerial.value||'').trim();
    if (!cfgId || !serial) { alert('请填写模板ID与序列号'); return; }
    try {
      const j = await apiPost(`/configs/${encodeURIComponent(cfgId)}/apply?serial=${encodeURIComponent(serial)}`);
      alert(j.message || '已应用');
    } catch (e) {
      alert('应用失败：' + e.message);
    }
  });

  els.loadAllBtn.addEventListener('click', async () => {
    els.debugOutput.textContent = '加载中...';
    try {
      const [servers, configs] = await Promise.all([
        apiGet('/servers?status=pending').catch(e=>({error:e.message})),
        apiGet('/configs').catch(e=>({error:e.message}))
      ]);
      els.debugOutput.textContent = JSON.stringify({servers, configs}, null, 2);
    } catch (e) {
      els.debugOutput.textContent = '失败：' + e.message;
    }
  });

  async function pingBackend() {
    const el = document.getElementById('backendStatus');
    if (!el) return;
    el.textContent = '后端状态：检测中…';
    el.classList.remove('ok','bad');
    try {
      const res = await fetch(API_BASE + '/health', { headers: headers() });
      if (res.ok) {
        el.textContent = '后端状态：已连接';
        el.classList.add('ok');
      } else {
        throw new Error('HTTP ' + res.status);
      }
    } catch (e) {
      el.textContent = '后端状态：未连接';
      el.classList.add('bad');
    }
  }

  // 初始化
  pingBackend();
  const statusSel = document.getElementById('serversStatus');
  const initStatus = statusSel ? statusSel.value : 'pending';
  loadServers(initStatus);
  loadConfigs();
})();
