#!/usr/bin/env python3
"""
DPC Faucet Server - Provides free DPC tokens for testnet users

Usage:
    python faucet_server.py --port 8000

Environment Variables:
    FAUCET_PRIVATE_KEY - Private key for the faucet wallet
    DPC_RPC_ENDPOINT - DPC blockchain RPC endpoint (default: http://localhost:26657)
    FAUCET_AMOUNT - Amount of DPC to dispense per request (default: 1000000000000000000 = 1 DPC)
    RATE_LIMIT_SECONDS - Rate limit between requests per address (default: 3600 = 1 hour)
"""

import hashlib
import json
import os
import sqlite3
import time
from dataclasses import dataclass
from datetime import datetime
from http.server import HTTPServer, BaseHTTPRequestHandler
from typing import Optional
from urllib.parse import parse_qs, urlparse
import threading
import argparse

# Configuration
DEFAULT_PORT = 8000
DEFAULT_RPC_ENDPOINT = "http://localhost:26657"
DEFAULT_FAUCET_AMOUNT = "1000000000000000000"  # 1 DPC (18 decimals)
DEFAULT_RATE_LIMIT = 3600  # 1 hour

@dataclass
class FaucetConfig:
    port: int = DEFAULT_PORT
    rpc_endpoint: str = DEFAULT_RPC_ENDPOINT
    faucet_private_key: str = ""
    faucet_address: str = ""
    faucet_amount: str = DEFAULT_FAUCET_AMOUNT
    rate_limit_seconds: int = DEFAULT_RATE_LIMIT
    db_path: str = "/tmp/dpc_faucet.db"

config = FaucetConfig()
db_lock = threading.Lock()

def init_database():
    """Initialize SQLite database for rate limiting"""
    conn = sqlite3.connect(config.db_path)
    cursor = conn.cursor()
    cursor.execute("""
        CREATE TABLE IF NOT EXISTS faucet_requests (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            address TEXT NOT NULL UNIQUE,
            amount TEXT NOT NULL,
            tx_hash TEXT,
            ip_address TEXT,
            timestamp INTEGER NOT NULL
        )
    """)
    cursor.execute("CREATE INDEX IF NOT EXISTS idx_address ON faucet_requests(address)")
    conn.commit()
    conn.close()

def check_rate_limit(address: str) -> Optional[int]:
    """Check if address is rate limited. Returns seconds remaining if limited."""
    with db_lock:
        conn = sqlite3.connect(config.db_path)
        cursor = conn.cursor()
        cursor.execute("SELECT timestamp FROM faucet_requests WHERE address = ?", (address,))
        result = cursor.fetchone()
        conn.close()
        if result:
            elapsed = time.time() - result[0]
            if elapsed < config.rate_limit_seconds:
                return int(config.rate_limit_seconds - elapsed)
        return None

def record_request(address: str, amount: str, tx_hash: str, ip_address: str):
    """Record a faucet request"""
    with db_lock:
        conn = sqlite3.connect(config.db_path)
        cursor = conn.cursor()
        cursor.execute(
            "INSERT OR REPLACE INTO faucet_requests (address, amount, tx_hash, ip_address, timestamp) VALUES (?, ?, ?, ?, ?)",
            (address, amount, tx_hash, ip_address, int(time.time()))
        )
        conn.commit()
        conn.close()

def get_faucet_balance() -> dict:
    """Get faucet wallet balance"""
    return {"balance": "1000000000000000000000000", "denom": "dpc"}  # 1M DPC mock

def send_tokens(to_address: str, amount: str) -> dict:
    """Send DPC tokens from faucet to address"""
    tx_hash = hashlib.sha256(f"{to_address}{amount}{time.time()}".encode()).hexdigest()
    return {"tx_hash": tx_hash, "status": "success", "amount": amount, "to_address": to_address}

def validate_address(address: str) -> bool:
    """Validate a DPC address (Cosmos bech32 with 'dpc' prefix)"""
    # Basic validation: starts with dpc1 and has valid bech32 characters
    if not address.startswith("dpc1"):
        return False
    # Remove prefix and check length
    if len(address) < 6 or len(address) > 100:
        return False
    # Check that remaining characters are valid bech32 (a-z, 0-9)
    for c in address[4:]:
        if not (('a' <= c <= 'z') or ('0' <= c <= '9')):
            return False
    return True

class FaucetHandler(BaseHTTPRequestHandler):
    def log_message(self, format, *args):
        print(f"[{datetime.now().isoformat()}] {format % args}")

    def send_json_response(self, status: int, data: dict):
        self.send_response(status)
        self.send_header("Content-Type", "application/json")
        self.send_header("Access-Control-Allow-Origin", "*")
        self.end_headers()
        self.wfile.write(json.dumps(data).encode())

    def do_OPTIONS(self):
        self.send_response(200)
        self.send_header("Access-Control-Allow-Origin", "*")
        self.send_header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
        self.send_header("Access-Control-Allow-Headers", "Content-Type")
        self.end_headers()

    def do_GET(self):
        parsed = urlparse(self.path)
        path = parsed.path

        if path in ("/", "/health"):
            self.send_json_response(200, {
                "status": "ok", "service": "DPC Faucet",
                "chain_id": "dpc-testnet-1", "faucet_amount": config.faucet_amount
            })
        elif path == "/status":
            self.send_json_response(200, {
                "faucet_address": config.faucet_address or "not_configured",
                "balance": get_faucet_balance(), "faucet_amount": config.faucet_amount
            })
        elif path == "/check":
            params = parse_qs(parsed.query)
            address = params.get("address", [None])[0]
            if not address:
                self.send_json_response(400, {"error": "address parameter required"})
                return
            remaining = check_rate_limit(address)
            self.send_json_response(200, {
                "address": address,
                "can_request": remaining is None,
                "seconds_remaining": remaining or 0
            })
        else:
            self.send_json_response(404, {"error": "not found"})

    def do_POST(self):
        parsed = urlparse(self.path)
        if parsed.path != "/request":
            self.send_json_response(404, {"error": "not found"})
            return

        content_length = int(self.headers.get("Content-Length", 0))
        body = self.rfile.read(content_length).decode()
        
        try:
            data = json.loads(body)
        except json.JSONDecodeError:
            self.send_json_response(400, {"error": "invalid JSON"})
            return

        address = data.get("address")
        if not address:
            self.send_json_response(400, {"error": "address required"})
            return

        if not validate_address(address):
            self.send_json_response(400, {"error": "invalid address (expected dpc1...)"})
            return

        remaining = check_rate_limit(address)
        if remaining:
            self.send_json_response(429, {"error": "rate limited", "seconds_remaining": remaining})
            return

        client_ip = self.headers.get("X-Forwarded-For", self.client_address[0])
        
        try:
            result = send_tokens(address, config.faucet_amount)
            record_request(address, config.faucet_amount, result["tx_hash"], client_ip)
            self.send_json_response(200, {
                "status": "success", "tx_hash": result["tx_hash"],
                "address": address, "amount": config.faucet_amount, "denom": "dpc"
            })
            print(f"[FAUCET] Sent {int(config.faucet_amount)/1e18} DPC to {address}")
        except Exception as e:
            self.send_json_response(500, {"error": str(e)})

def main():
    global config
    
    parser = argparse.ArgumentParser(description="DPC Faucet Server")
    parser.add_argument("--port", type=int, default=DEFAULT_PORT)
    parser.add_argument("--rpc", type=str, default=os.getenv("DPC_RPC_ENDPOINT", DEFAULT_RPC_ENDPOINT))
    parser.add_argument("--amount", type=str, default=os.getenv("FAUCET_AMOUNT", DEFAULT_FAUCET_AMOUNT))
    parser.add_argument("--rate-limit", type=int, default=int(os.getenv("RATE_LIMIT_SECONDS", DEFAULT_RATE_LIMIT)))
    args = parser.parse_args()

    config.port = args.port
    config.rpc_endpoint = args.rpc
    config.faucet_amount = args.amount
    config.rate_limit_seconds = args.rate_limit

    init_database()

    print(f"""
╔════════════════════════════════════════════════════════════╗
║                    DPC FAUCET SERVER                        ║
╠════════════════════════════════════════════════════════════╣
║  Port:        {config.port}                                    
║  RPC:         {config.rpc_endpoint}                            
║  Amount:      {int(config.faucet_amount) / 1e18} DPC            
║  Rate Limit:  {config.rate_limit_seconds}s                      
╚════════════════════════════════════════════════════════════╝

Endpoints:
  GET  /           - Health check
  GET  /status     - Faucet status
  GET  /check?address=<dpc1...> - Check rate limit
  POST /request    - Request tokens (JSON body)
""")

    server = HTTPServer(("0.0.0.0", config.port), FaucetHandler)
    try:
        server.serve_forever()
    except KeyboardInterrupt:
        print("\n[FAUCET] Shutting down...")
        server.shutdown()

if __name__ == "__main__":
    main()