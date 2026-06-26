#!/usr/bin/env python3
import os
import sys
import unittest

SCRIPTS = __import__("pathlib").Path(__file__).resolve().parent
sys.path.insert(0, str(SCRIPTS))

from reading_llm_client import _llm_log, _port_busy_quick_sec  # noqa: E402


class ReadingLLMClientLogTest(unittest.TestCase):
    def test_port_busy_quick_default(self):
        os.environ.pop("LLAMACPP_PORT_BUSY_QUICK_SEC", None)
        self.assertEqual(_port_busy_quick_sec(), 3)

    def test_port_busy_quick_env(self):
        os.environ["LLAMACPP_PORT_BUSY_QUICK_SEC"] = "5"
        try:
            self.assertEqual(_port_busy_quick_sec(), 5)
        finally:
            os.environ.pop("LLAMACPP_PORT_BUSY_QUICK_SEC", None)

    def test_llm_log_has_timestamp(self):
        import io
        from contextlib import redirect_stdout

        buf = io.StringIO()
        with redirect_stdout(buf):
            _llm_log("reading-llm", "hello")
        line = buf.getvalue().strip()
        self.assertRegex(line, r"^\[reading-llm \d{2}:\d{2}:\d{2}\] hello$")


if __name__ == "__main__":
    unittest.main()
