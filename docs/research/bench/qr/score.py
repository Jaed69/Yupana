import json, sys
from collections import defaultdict
truth = json.load(open(sys.argv[1]))
runs = {}
times = {}
for p in sys.argv[2:]:
    d = json.load(open(p)); runs.update(d["results"]); times.update(d["ms_per_image"])
levels = sorted({v["level"] for v in truth.values()})
print("decoder".ljust(30), " ".join(f"L{l}".rjust(6) for l in levels), "total".rjust(7), "wrong".rjust(6), "ms/img".rjust(7))
for name, res in sorted(runs.items()):
    ok = defaultdict(int); n = defaultdict(int); wrong = 0
    for f, t in truth.items():
        n[t["level"]] += 1
        got = res.get(f)
        if got == t["text"]: ok[t["level"]] += 1
        elif got: wrong += 1
    tot = sum(ok.values())
    print(name.ljust(30), " ".join(f"{ok[l]}/{n[l]}".rjust(6) for l in levels), f"{tot}/{len(truth)}".rjust(7), str(wrong).rjust(6), f"{times[name]:.1f}".rjust(7))
# by QR size at the hardest level
for l in levels[-2:]:
    print(f"\nLevel {l} by qr_px:")
    for name, res in sorted(runs.items()):
        row = []
        for px in sorted({v['qr_px'] for v in truth.values()}):
            fs = [f for f, t in truth.items() if t["level"] == l and t["qr_px"] == px]
            row.append(f"{px}px {sum(res.get(f) == truth[f]['text'] for f in fs)}/{len(fs)}")
        print(" ", name.ljust(30), "  ".join(row))
