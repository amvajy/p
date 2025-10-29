import json
from http.server import HTTPServer, BaseHTTPRequestHandler
from urllib.parse import urlparse, parse_qs

AUDIT_LOG_PATH = './logs/audit.log'

# 内存数据模拟
servers = [
    {"serial": "ABC123", "hostname": "srv-abc", "ipAddress": "192.168.88.10", "macAddress": "00:11:22:33:44:55", "status": "pending"},
    {"serial": "XYZ789", "hostname": "srv-xyz", "ipAddress": "192.168.88.11", "macAddress": "66:77:88:99:AA:BB", "status": "pending"},
]
configs = [
    {"id": 1, "name": "CentOS7-Base", "systemType": "CentOS", "systemVersion": "7", "description": "Base install", "configContent": "#kickstart", "kernelParams": "text", "packages": "vim,net-tools"},
    {"id": 2, "name": "Ubuntu20-Base", "systemType": "Ubuntu", "systemVersion": "20.04", "description": "Base install", "configContent": "#preseed", "kernelParams": "auto", "packages": "curl,wget"},
]
next_cfg_id = 3

class Handler(BaseHTTPRequestHandler):
    def _set_headers(self, code=200, content_type='application/json'):
        self.send_response(code)
        self.send_header('Content-Type', content_type)
        self.send_header('Access-Control-Allow-Origin', '*')
        self.send_header('Access-Control-Allow-Methods', 'GET,POST,PUT,DELETE,OPTIONS')
        self.send_header('Access-Control-Allow-Headers', 'Authorization, Content-Type')
        self.end_headers()

    def do_OPTIONS(self):
        self._set_headers(204)

    def do_GET(self):
        parsed = urlparse(self.path)
        path = parsed.path
        qs = parse_qs(parsed.query)
        # health
        if path == '/api/health':
            self._set_headers(200)
            self.wfile.write(json.dumps({'status':'ok'}).encode('utf-8'))
            return
        # audit logs
        if path == '/api/audit/logs':
            try:
                logs = []
                with open(AUDIT_LOG_PATH, 'r', encoding='utf-8') as f:
                    for line in f:
                        try:
                            logs.append(json.loads(line))
                        except Exception:
                            pass
                order = (qs.get('order',["desc"]) or ["desc"])[0]
                offset = int((qs.get('offset',["0"]) or ["0"])[0])
                limit = int((qs.get('limit',["100"]) or ["100"])[0])
                if order == 'desc':
                    logs = list(reversed(logs))
                if offset < 0:
                    offset = 0
                if limit <= 0:
                    limit = 100
                logs = logs[offset:offset+limit]
                self._set_headers(200)
                self.wfile.write(json.dumps(logs).encode('utf-8'))
            except FileNotFoundError:
                self._set_headers(200)
                self.wfile.write(b"[]")
            return
        # servers list
        if path == '/api/servers':
            status = (qs.get('status',["pending"]) or ["pending"])[0]
            out = [s for s in servers if (status=='' or s.get('status')==status)]
            self._set_headers(200)
            self.wfile.write(json.dumps(out).encode('utf-8'))
            return
        # server detail
        if path.startswith('/api/servers/'):
            serial = path.split('/')[-1]
            s = next((x for x in servers if x.get('serial')==serial), None)
            if not s:
                self._set_headers(404)
                self.wfile.write(json.dumps({'error':'not found'}).encode('utf-8'))
            else:
                self._set_headers(200)
                self.wfile.write(json.dumps(s).encode('utf-8'))
            return
        # configs list
        if path == '/api/configs':
            self._set_headers(200)
            self.wfile.write(json.dumps([{
                'id': c['id'], 'name': c['name'], 'systemType': c['systemType'], 'systemVersion': c['systemVersion']
            } for c in configs]).encode('utf-8'))
            return
        # config detail
        if path.startswith('/api/configs/'):
            try:
                cid = int(path.split('/')[-1])
            except Exception:
                self._set_headers(400)
                self.wfile.write(json.dumps({'error':'bad id'}).encode('utf-8'))
                return
            c = next((x for x in configs if x['id']==cid), None)
            if not c:
                self._set_headers(404)
                self.wfile.write(json.dumps({'error':'not found'}).encode('utf-8'))
            else:
                self._set_headers(200)
                self.wfile.write(json.dumps(c).encode('utf-8'))
            return
        # default
        self._set_headers(404)
        self.wfile.write(json.dumps({'error':'not found'}).encode('utf-8'))

    def do_POST(self):
        parsed = urlparse(self.path)
        path = parsed.path
        qs = parse_qs(parsed.query)
        # confirm/install
        if path.startswith('/api/servers/') and (path.endswith('/confirm') or path.endswith('/install')):
            parts = path.split('/')
            serial = parts[3]
            s = next((x for x in servers if x.get('serial')==serial), None)
            if not s:
                self._set_headers(404)
                self.wfile.write(json.dumps({'error':'not found'}).encode('utf-8'))
                return
            if path.endswith('/confirm'):
                s['status'] = 'confirmed'
                self._set_headers(200)
                self.wfile.write(json.dumps({'message':'已确认'}).encode('utf-8'))
                return
            if path.endswith('/install'):
                s['status'] = 'installed'
                self._set_headers(200)
                self.wfile.write(json.dumps({'message':'已标记安装'}).encode('utf-8'))
                return
        # create config
        if path == '/api/configs':
            length = int(self.headers.get('Content-Length','0'))
            body = self.rfile.read(length) if length>0 else b''
            try:
                payload = json.loads(body.decode('utf-8')) if body else {}
            except Exception:
                payload = {}
            global next_cfg_id
            c = {
                'id': next_cfg_id,
                'name': payload.get('name',''),
                'description': payload.get('description',''),
                'systemType': payload.get('systemType','CentOS'),
                'systemVersion': payload.get('systemVersion',''),
                'configContent': payload.get('configContent',''),
                'kernelParams': payload.get('kernelParams',''),
                'packages': payload.get('packages',''),
            }
            configs.append(c)
            next_cfg_id += 1
            self._set_headers(200)
            self.wfile.write(json.dumps({'id': c['id']}).encode('utf-8'))
            return
        # apply config
        if path.startswith('/api/configs/') and path.endswith('/apply'):
            cid = int(path.split('/')[-2])
            serial = (qs.get('serial',[""]) or [""])[0]
            self._set_headers(200)
            self.wfile.write(json.dumps({'message': f'模板 {cid} 已应用到 {serial}'}).encode('utf-8'))
            return
        # default
        self._set_headers(404)
        self.wfile.write(json.dumps({'error':'not found'}).encode('utf-8'))

    def do_PUT(self):
        parsed = urlparse(self.path)
        path = parsed.path
        if path.startswith('/api/configs/'):
            try:
                cid = int(path.split('/')[-1])
            except Exception:
                self._set_headers(400)
                self.wfile.write(json.dumps({'error':'bad id'}).encode('utf-8'))
                return
            length = int(self.headers.get('Content-Length','0'))
            body = self.rfile.read(length) if length>0 else b''
            try:
                payload = json.loads(body.decode('utf-8')) if body else {}
            except Exception:
                payload = {}
            c = next((x for x in configs if x['id']==cid), None)
            if not c:
                self._set_headers(404)
                self.wfile.write(json.dumps({'error':'not found'}).encode('utf-8'))
                return
            # 简单覆盖更新
            for k in ['name','description','systemType','systemVersion','configContent','kernelParams','packages']:
                if k in payload:
                    c[k] = payload[k]
            self._set_headers(200)
            self.wfile.write(json.dumps({'message':'updated'}).encode('utf-8'))
            return
        self._set_headers(404)
        self.wfile.write(json.dumps({'error':'not found'}).encode('utf-8'))

if __name__ == '__main__':
    server = HTTPServer(('0.0.0.0', 8080), Handler)
    print('Dev API stub listening on http://localhost:8080')
    server.serve_forever()
