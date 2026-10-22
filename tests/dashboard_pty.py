#!/usr/bin/env python3
"""VT screen assertions against the built CLI and a local, synthetic API fixture."""
import codecs
import fcntl
import http.server
import json
import os
from pathlib import Path
import pty
import select
import signal
import struct
import subprocess
import tempfile
import termios
import threading
import time
import pyte

ROOT = Path(__file__).resolve().parents[1]
BINARY = ROOT / 'bin' / 'SINet'


class API(http.server.BaseHTTPRequestHandler):
    def do_GET(self):
        assert self.path == '/v1/me', self.path
        data = b'{"handle":"fixture"}'
        self.send_response(200)
        self.send_header('Content-Type', 'application/json')
        self.end_headers()
        self.wfile.write(data)

    def log_message(self, *args):
        pass


class Terminal:
    def __init__(self, env, args=(), width=80, height=24):
        self.master, self.slave = pty.openpty()
        self.original = termios.tcgetattr(self.slave)
        fcntl.ioctl(self.slave, termios.TIOCSWINSZ, struct.pack('HHHH', height, width, 0, 0))
        self.screen = pyte.Screen(width, height)
        self.stream = pyte.Stream(self.screen)
        self.decoder = codecs.getincrementaldecoder('utf-8')('replace')
        self.raw = b''
        self.proc = subprocess.Popen([str(BINARY), *args], stdin=self.slave, stdout=self.slave,
                                     stderr=self.slave, env=env)

    def pump(self, duration=0.2):
        end = time.monotonic() + duration
        while time.monotonic() < end:
            if select.select([self.master], [], [], 0.02)[0]:
                chunk = os.read(self.master, 65536)
                self.raw += chunk
                self.stream.feed(self.decoder.decode(chunk))

    def expect(self, text):
        end = time.monotonic() + 10
        while time.monotonic() < end:
            self.pump(0.1)
            if text in '\n'.join(self.screen.display):
                return
        raise AssertionError(f'missing {text!r}\n' + '\n'.join(self.screen.display))

    def send(self, data):
        os.write(self.master, data.encode())
        self.pump()

    def close(self):
        try:
            if self.proc.poll() is None:
                self.send('q')
                self.proc.wait(timeout=5)
            assert self.proc.returncode == 0, self.raw[-2000:]
            assert termios.tcgetattr(self.slave) == self.original, 'terminal not restored'
        finally:
            if self.proc.poll() is None:
                self.proc.kill()
                self.proc.wait()
            os.close(self.master)
            os.close(self.slave)


def main():
    with tempfile.TemporaryDirectory(prefix='sinet-ui-') as tmp:
        api = http.server.HTTPServer(('127.0.0.1', 0), API)
        threading.Thread(target=api.serve_forever, daemon=True).start()
        config = Path(tmp) / 'session.json'
        config.write_text(json.dumps({'api': f'http://127.0.0.1:{api.server_port}', 'token': 'a'*64}))
        config.chmod(0o600)
        env = dict(os.environ, TERM='xterm-256color', SINET_CONFIG=str(config),
                   SINET_API=f'http://127.0.0.1:{api.server_port}', SINET_UI_CONFIG=tmp+'/ui.json')
        env.pop('NO_COLOR', None)
        env['COLORTERM'] = 'truecolor'
        t = Terminal(env)
        try:
            t.expect('8  Wallet, Email and Services')
            initial_title_row = next(i for i,s in enumerate(t.screen.display) if 'Your network' in s)
            for _ in range(18):
                # Fragment the escape sequence across separate reads, as on SSH.
                os.write(t.master, b'\x1b')
                time.sleep(0.005)
                os.write(t.master, b'[B')
                t.pump(0.03)
            t.pump()
            assert sum('Your network' in s for s in t.screen.display) == 1, repr(t.raw[-4000:]) + '\n' + '\n'.join(t.screen.display)
            assert 'Your network' in t.screen.display[initial_title_row], 'navigation scrolled menu'
            assert any('› 1  Profile' in s for s in t.screen.display)
            t.send('\x1b[200~8\n\x1b[201~')
            assert any('Your network' in s for s in t.screen.display), 'paste triggered an action'
            t.send('8')
            t.expect('Wallet, email and services')
            t.expect('2  Email')
            t.send('9')
            t.expect('Your network')
            t.send('7')
            t.expect('Settings')
            t.send('1')
            t.expect('Theme: dark')
            t.send('1')
            t.expect('Light · signal blue')
            t.send('2')
            t.expect('Theme: light')
            saved = json.loads(Path(env['SINET_UI_CONFIG']).read_text())
            assert saved['theme'] == 'light'
            t.send('q')
            t.expect('Show account settings')
            t.send('q')
            t.expect('Your network')
            assert b'\x1b[38;' in t.raw or b'\x1b[9' in t.raw, 'colour missing'
            fcntl.ioctl(t.slave, termios.TIOCSWINSZ, struct.pack('HHHH', 14, 40, 0, 0))
            t.screen.resize(lines=14, columns=40)
            t.proc.send_signal(signal.SIGWINCH)
            t.pump()
            t.send('k')
            t.expect('› 9  Exit')
        finally:
            t.close()
            api.shutdown()
        # NO_COLOR wins over a saved preference and works with no API/session.
        env['NO_COLOR'] = '1'
        env['SINET_API'] = 'invalid'
        Path(env['SINET_UI_CONFIG']).write_text(json.dumps({'theme':'dark','color':'always','density':'compact'}))
        t = Terminal(env, ['ui'])
        try:
            t.expect('Colour: always')
            assert b'\x1b[38;' not in t.raw and b'\x1b[48;' not in t.raw
        finally:
            t.close()
        plain = subprocess.run([str(BINARY), 'ui'], input='q\n', text=True, capture_output=True, env=env, check=True)
        assert '\x1b' not in plain.stdout and '1. Theme:' in plain.stdout
    print('PASS dashboard PTY: stable rows, fragmented arrows, all sections, paste safety, appearance persistence, resize, colour/NO_COLOR, redirected output, terminal restore (synthetic API)')


if __name__ == '__main__':
    main()
