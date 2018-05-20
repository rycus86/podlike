import re
from BaseHTTPServer import HTTPServer, BaseHTTPRequestHandler

class Handler(BaseHTTPRequestHandler):
    def do_GET(self):
        result = reduce(
            lambda x, y: x * y, (int(x) for x in self.path.split('/') if re.match('^-?[0-9]+$', x))
        )

        self.send_response(200)
        self.end_headers()
        self.wfile.write('%d\n' % result)

HTTPServer(('0.0.0.0', 5000), Handler).serve_forever()