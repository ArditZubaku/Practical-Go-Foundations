// Bun's built-in HTTP server

Bun.serve({
  port: 8080,
  fetch(req) {
    const url = new URL(req.url);

    // Exact match for "/health"
    if (url.pathname === "/health" && req.method === "GET") {
      return new Response("OK\n", {
        status: 200,
        headers: {
          "Content-Type": "text/plain"
        }
      });
    }

    // Default 404
    return new Response("Not Found\n", {
      status: 404,
      headers: {
        "Content-Type": "text/plain"
      }
    });
  }
});

console.log("Server running at http://localhost:8080");

