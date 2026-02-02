import os
import subprocess
import signal
import sys
import threading
import time
from pathlib import Path

CACHE_DIR = Path.home() / ".cache" / "jtvt"
AUDIO_PATH = CACHE_DIR / "dict.wav"
PID_PATH = CACHE_DIR / "rec.pid"

# Change these defaults if you want:
DATA_DIR = Path.home() / ".local" / "share" / "jtt"
WHISPER_MODEL = DATA_DIR / "ggml-small.en.bin"
OLLAMA_MODEL = "llama3.2:3b"

def _run(cmd: list[str], check: bool = True) -> subprocess.CompletedProcess:
    return subprocess.run(cmd, check=check, text=True, capture_output=True)

class Spinner:
    def __init__(self):
        self._stop = threading.Event()
        self._thread = None

    def __enter__(self):
        self._stop.clear()
        self._thread = threading.Thread(target=self._spin, daemon=True)
        self._thread.start()
        return self

    def __exit__(self, *args):
        self._stop.set()
        self._thread.join()
        sys.stdout.write("\r\033[K")
        sys.stdout.flush()

    def _spin(self):
        chars = "⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏"
        i = 0
        while not self._stop.is_set():
            sys.stdout.write(f"\r{chars[i % len(chars)]}")
            sys.stdout.flush()
            i += 1
            time.sleep(0.08)

def start() -> int:
    CACHE_DIR.mkdir(parents=True, exist_ok=True)
    if PID_PATH.exists():
        return 0  # already recording; idempotent

    # Start SoX recording in background
    # rec -q -c 1 -r 16000 dict.wav trim 0 600
    p = subprocess.Popen(
        ["rec", "-q", "-c", "1", "-r", "16000", str(AUDIO_PATH), "trim", "0", "600"],
        stdout=subprocess.DEVNULL,
        stderr=subprocess.DEVNULL,
    )
    PID_PATH.write_text(str(p.pid))
    return 0

def stop(debug: bool = False) -> int:
    if not PID_PATH.exists():
        raise SystemExit("No active recording (missing rec.pid).")

    pid = int(PID_PATH.read_text().strip())
    PID_PATH.unlink(missing_ok=True)

    # Stop recorder
    try:
        os.kill(pid, signal.SIGTERM)
    except ProcessLookupError:
        pass

    time.sleep(0.1)
    if not AUDIO_PATH.exists():
        raise SystemExit("Missing audio file; recording may have failed.")

    if not WHISPER_MODEL.exists():
        raise SystemExit(f"Whisper model not found: {WHISPER_MODEL}")

    # Transcribe
    t0 = time.time()
    ctx = Spinner() if not debug else None
    if ctx:
        ctx.__enter__()
    try:
        _run([
            "whisper-cli",
            "-m", str(WHISPER_MODEL),
            "-f", str(AUDIO_PATH),
            "--no-timestamps",
            "--language", "en",
            "--output-txt",
            "--output-file", str(AUDIO_PATH.with_suffix("")),
        ])
    finally:
        if ctx:
            ctx.__exit__(None, None, None)
    whisper_time = time.time() - t0

    raw_txt = AUDIO_PATH.with_suffix(".txt")
    if not raw_txt.exists():
        raise SystemExit("Missing transcript output from whisper-cpp.")

    transcript = raw_txt.read_text().strip()
    if debug:
        print(f"(whisper {whisper_time:.1f}s) {transcript}")

    prompt = (
        "Clean this voice transcript. Output ONLY the cleaned text, nothing else.\n"
        "Rules: remove filler words (um, uh, like), fix punctuation and casing, keep original wording.\n"
        "Transcript:\n"
    )

    # Clean with Ollama
    t0 = time.time()
    if ctx:
        ctx.__enter__()
    try:
        proc = subprocess.run(
            ["ollama", "run", OLLAMA_MODEL],
            input=prompt + "\n" + transcript,
            text=True,
            capture_output=True,
            check=True,
        )
    finally:
        if ctx:
            ctx.__exit__(None, None, None)
    ollama_time = time.time() - t0
    cleaned = proc.stdout.strip()

    if debug:
        print(f"(ollama {ollama_time:.1f}s) {cleaned}")

    # Copy to clipboard
    subprocess.run(["pbcopy"], input=cleaned, text=True, check=True)
    return 0

def main() -> None:
    import argparse

    parser = argparse.ArgumentParser(prog="jtt")
    sub = parser.add_subparsers(dest="cmd", required=True)

    sub.add_parser("start")
    stop_parser = sub.add_parser("stop")
    stop_parser.add_argument("--debug", action="store_true", help="print timings and intermediate outputs")

    args = parser.parse_args()
    if args.cmd == "start":
        raise SystemExit(start())
    if args.cmd == "stop":
        raise SystemExit(stop(debug=args.debug))
