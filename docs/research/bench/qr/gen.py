"""Generate synthetic 'phone photo of a paper invoice with SUNAT QR' images with controlled degradations."""
import json, random, sys
from pathlib import Path
import numpy as np, cv2, qrcode

OUT = Path(sys.argv[1]); OUT.mkdir(parents=True, exist_ok=True)
rng = random.Random(42); nrng = np.random.default_rng(42)

def payload(i):
    ruc = f"20{rng.randrange(10**8, 10**9)}"
    return f"{ruc}|01|F{rng.randrange(1,999):03d}|{rng.randrange(1,99999999):08d}|{rng.uniform(1,500):.2f}|{rng.uniform(10,3000):.2f}|2026-09-{rng.randrange(1,29):02d}|6|20{rng.randrange(10**8,10**9)}|{''.join(rng.choice('ABCDEFabcdef0123456789+/=') for _ in range(28))}|"

def invoice(text, qr_px, ec):
    q = qrcode.QRCode(error_correction=ec, box_size=1, border=4); q.add_data(text); q.make()
    m = np.array(q.get_matrix(), dtype=np.uint8)
    qr = np.where(m, 0, 255).astype(np.uint8)
    qr = cv2.resize(qr, (qr_px, qr_px), interpolation=cv2.INTER_NEAREST)
    page = np.full((1400, 1000), 245, np.uint8)
    for y in range(80, 1000, 28):  # fake printed text lines
        x2 = rng.randrange(400, 920); cv2.line(page, (60, y), (x2, y), 40, 3)
    y0, x0 = 1050, 1000 // 2 - qr_px // 2
    page[y0:y0 + qr_px, x0:x0 + qr_px] = qr
    return cv2.cvtColor(page, cv2.COLOR_GRAY2BGR)

def photo(img, lvl):
    h, w = img.shape[:2]
    # perspective (phone held at an angle)
    d = int(w * 0.04 * lvl)
    src = np.float32([[0, 0], [w, 0], [w, h], [0, h]])
    dst = np.float32([[rng.randint(0, d), rng.randint(0, d)], [w - rng.randint(0, d), rng.randint(0, d)],
                      [w - rng.randint(0, d), h - rng.randint(0, d)], [rng.randint(0, d), h - rng.randint(0, d)]])
    img = cv2.warpPerspective(img, cv2.getPerspectiveTransform(src, dst), (w, h), borderValue=(90, 80, 70))
    # rotation
    M = cv2.getRotationMatrix2D((w / 2, h / 2), rng.uniform(-8, 8) * lvl, 1.0)
    img = cv2.warpAffine(img, M, (w, h), borderValue=(90, 80, 70))
    # uneven lighting / shadow
    gx = np.linspace(1.0, 1.0 - 0.18 * lvl, w)[None, :, None]
    img = np.clip(img * gx * rng.uniform(1 - 0.12 * lvl, 1.0), 0, 255).astype(np.uint8)
    # blur (defocus + motion)
    if lvl > 0: img = cv2.GaussianBlur(img, (0, 0), 0.9 * lvl)
    if lvl >= 1.5:
        mk = np.zeros((int(3 * lvl), int(3 * lvl))); mk[mk.shape[0] // 2, :] = 1 / mk.shape[1]
        img = cv2.filter2D(img, -1, mk)
    # sensor noise
    img = np.clip(img + nrng.normal(0, 4 * lvl, img.shape), 0, 255).astype(np.uint8)
    # downscale like a messaging-app recompression, then JPEG
    s = 1.0 - 0.12 * lvl
    img = cv2.resize(img, (int(w * s), int(h * s)), interpolation=cv2.INTER_AREA)
    ok, buf = cv2.imencode(".jpg", img, [cv2.IMWRITE_JPEG_QUALITY, int(90 - 12 * lvl)])
    return buf

truth = {}
ecs = {"M": qrcode.constants.ERROR_CORRECT_M, "Q": qrcode.constants.ERROR_CORRECT_Q}
i = 0
for lvl in [0, 0.5, 1.0, 1.25, 1.5, 1.75, 2.0]:
    for qr_px in [120, 160, 220]:
        for ec_name, ec in ecs.items():
            for rep in range(6):
                text = payload(i)
                name = f"l{lvl:.2f}_q{qr_px}_{ec_name}_{rep}.jpg"
                (OUT / name).write_bytes(photo(invoice(text, qr_px, ec), lvl).tobytes())
                truth[name] = {"text": text, "level": lvl, "qr_px": qr_px, "ec": ec_name}
                i += 1
(OUT / "truth.json").write_text(json.dumps(truth, indent=1))
print(len(truth), "images")
