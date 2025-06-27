const http = require('http');
const url = require('url');

const PORT = 8080;

// Health handler function
function healthHandler(req, res) {
  // TODO: Run a health check
  res.writeHead(200, { 'Content-Type': 'text/plain' });
  res.end('OK\n');
}

// Create the server
const server = http.createServer((req, res) => {
  const parsedUrl = url.parse(req.url, true);

  // Exact route match
  if (req.method === 'GET' && parsedUrl.pathname === '/health') {
    return healthHandler(req, res);
  }

  // Fallback 404
  res.writeHead(404, { 'Content-Type': 'text/plain' });
  res.end('Not Found\n');
});

// Start the server
server.listen(PORT, () => {
  console.log(`Server listening on http://localhost:${PORT}`);
});

