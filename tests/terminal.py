"""Small PTY adapter: POSIX terminals and Windows' real ConPTY."""
import os
import select
import subprocess
import time


class Terminal:
    def __init__(self, binary, env):
        self.windows = os.name == "nt"
        if self.windows:
            from winpty import PtyProcess, Backend
            self.proc = PtyProcess.spawn([binary], env=env, dimensions=(24, 100),
                                         backend=Backend.ConPTY)
            self.reader = self.proc.fileobj
        else:
            import pty
            self.master, slave = pty.openpty()
            self.proc = subprocess.Popen([binary], stdin=slave, stdout=slave,
                                         stderr=slave, env=env)
            os.close(slave)
            self.reader = self.master

    def read(self):
        ready, _, _ = select.select([self.reader], [], [], 0.2)
        if not ready:
            return b""
        if self.windows:
            return self.reader.recv(8192)
        return os.read(self.master, 8192)

    def send(self, text):
        if self.windows:
            self.proc.write(text.replace("\n", "\r"))
        else:
            os.write(self.master, text.encode())

    def wait(self):
        if not self.windows:
            return self.proc.wait(timeout=10)
        deadline = time.monotonic() + 10
        while self.proc.isalive() and time.monotonic() < deadline:
            self.read()  # Drain ConPTY so child shutdown cannot block on output.
        if self.proc.isalive():
            raise AssertionError("ConPTY child failed to exit")
        return self.proc.exitstatus

    def close(self):
        if self.windows:
            self.proc.close(force=True)
        else:
            if self.proc.poll() is None:
                self.proc.kill()
                self.proc.wait()
            os.close(self.master)
