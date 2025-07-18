#!/bin/bash
# Test what headers are being sent

echo "Starting test server to capture headers..."

# Simple HTTP server that logs all headers
python3 -c '
import http.server
import socketserver

class HeaderLogger(http.server.BaseHTTPRequestHandler):
    def do_POST(self):
        print("\n=== REQUEST RECEIVED ===")
        print(f"Path: {self.path}")
        print("\nHeaders:")
        for header, value in self.headers.items():
            print(f"  {header}: {value}")
        
        # Read body
        content_length = int(self.headers.get("Content-Length", 0))
        body = self.rfile.read(content_length).decode("utf-8")
        print(f"\nBody: {body}")
        print("=======================\n")
        
        # Send response
        self.send_response(200)
        self.send_header("Content-Type", "application/json")
        self.end_headers()
        self.wfile.write(b"{\"error\": \"This is just a test server\"}")

with socketserver.TCPServer(("", 8080), HeaderLogger) as httpd:
    print("Test server listening on port 8080...")
    print("Try creating a todo from Claude Code to see what headers it sends")
    httpd.serve_forever()
'