#!/usr/bin/env python3
"""Enkel mock av en OpenAI-liknande leverantör.

Agentens definerare och proposerare pratar med en leverantör via ett
NDJSON-flöde. Mocken svarar med ett bestämt svar, så att Go- och
Rust-agenten kan köras mot samma "modell" och deras beteende jämföras.

Användning:
    scripts/mock-llm.py [port] [svar]

Svaret kan också läsas ur filen MOCK_RESPONSE_FILE. Utan svar returneras
en enkel definition.
"""

import json
import os
import sys
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer

STANDARD_SVAR = '{"title":"Mockmål","description":"Ett mål från mocken","icon":"🎯"}'


def läs_svar():
    """Läser svaret mocken ska ge."""
    fil = os.environ.get("MOCK_RESPONSE_FILE")

    if fil and os.path.exists(fil):
        with open(fil, encoding="utf-8") as fh:
            return fh.read()

    if len(sys.argv) > 2:
        return sys.argv[2]

    return os.environ.get("MOCK_RESPONSE", STANDARD_SVAR)


class Handler(BaseHTTPRequestHandler):
    """Svarar på varje POST med ett NDJSON-flöde."""

    def do_POST(self):
        """Läser requesten och svarar med modellens svar."""
        längd = int(self.headers.get("Content-Length", 0))

        if längd:
            self.rfile.read(längd)

        svar = läs_svar()

        rader = [
            json.dumps(
                {"choices": [{"index": 0, "delta": {"content": svar}}]},
                ensure_ascii=False,
            ),
            json.dumps({"choices": [{"index": 0, "delta": {}, "finish_reason": "stop"}]}),
            json.dumps(
                {
                    "usage": {
                        "prompt_tokens": 10,
                        "completion_tokens": 5,
                        "total_tokens": 15,
                    }
                }
            ),
        ]

        kropp = ("\n".join(rader) + "\n").encode("utf-8")

        self.send_response(200)
        self.send_header("Content-Type", "application/x-ndjson")
        self.send_header("Content-Length", str(len(kropp)))
        self.end_headers()
        self.wfile.write(kropp)

    def log_message(self, format, *args):
        """Håller tyst - loggen fylls bara av arbetsprocessen."""
        return


def main():
    port = int(sys.argv[1]) if len(sys.argv) > 1 else 8299

    print(f"Mockar leverantören på port {port}", flush=True)

    ThreadingHTTPServer(("127.0.0.1", port), Handler).serve_forever()


if __name__ == "__main__":
    main()
