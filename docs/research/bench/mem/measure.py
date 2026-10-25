import subprocess, sys, time, urllib.request, os
from pathlib import Path
import psutil

D = Path(__file__).parent
PY = str(D / ".venv" / "Scripts" / "python.exe")
cases = [
    ("fastapi + deps w/o QR stack", [PY, "-m", "uvicorn", "app_noqr:app", "--port", "8104"], 8104),
    ("go net/http + anthropic + pgx", [str(D / "gosvc.exe")]  # build: CGO_ENABLED=0 go build -ldflags="-s -w" -o gosvc.exe ., 8101),
    ("fastapi minimal (1 worker)", [PY, "-m", "uvicorn", "app_min:app", "--port", "8102"], 8102),
    ("fastapi + full deps (1 worker)", [PY, "-m", "uvicorn", "app_full:app", "--port", "8103"], 8103),
]
for name, cmd, port in cases:
    p = subprocess.Popen(cmd, cwd=D, env={**os.environ, "PORT": str(port)},
                         stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    url = f"http://127.0.0.1:{port}/"
    for _ in range(100):
        try: urllib.request.urlopen(url, timeout=1).read(); break
        except Exception: time.sleep(0.2)
    for _ in range(2000): urllib.request.urlopen(url, timeout=2).read()
    time.sleep(0.5)
    procs = [psutil.Process(p.pid)] + psutil.Process(p.pid).children(recursive=True)
    rss = sum(q.memory_info().rss for q in procs) / 2**20
    uss = sum(q.memory_full_info().uss for q in procs) / 2**20
    print(f"{name:34} RSS={rss:6.1f} MB  USS={uss:6.1f} MB")
    for q in procs: q.kill()
