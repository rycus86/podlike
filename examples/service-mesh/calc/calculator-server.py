import urllib2
from BaseHTTPServer import HTTPServer, BaseHTTPRequestHandler

class Handler(BaseHTTPRequestHandler):
    def do_GET(self):
        parts = [p for p in self.path.split('/') if p]
        op = parts[0]

        if op == 'add':
            result = urllib2.urlopen(
                urllib2.Request(
                    'http://localhost/v2/add/%s' % '/'.join(parts[1:]),
                    headers=self.headers)
            ).read()
        elif op == 'mul':
            result = urllib2.urlopen(
                urllib2.Request(
                    'http://localhost/v2/mul/%s' % '/'.join(parts[1:]),
                    headers=self.headers)
            ).read()
        elif op == 'sub':
            result = int(parts[1]) - int(parts[2])
        elif op == 'div':
            result = float(parts[1]) / float(parts[2])
        else:
            result = 'unknown'

        self.send_response(200)
        self.end_headers()
        self.wfile.write('%s\n' % str(result).strip())

HTTPServer(('0.0.0.0', 5000), Handler).serve_forever()