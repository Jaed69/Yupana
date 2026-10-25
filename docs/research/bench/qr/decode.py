"""Decode every image with several Python decoders; print JSON {decoder: {file: text|null}}."""
import json, sys, time
from pathlib import Path
import cv2, zxingcpp
from pyzbar import pyzbar

D = Path(sys.argv[1])
files = sorted(p for p in D.glob("*.jpg"))
det = cv2.QRCodeDetector()
try:
    wc = cv2.wechat_qrcode_WeChatQRCode()
except Exception as e:
    wc = None; print("wechat unavailable:", e, file=sys.stderr)

def opencv(img):
    t, _, _ = det.detectAndDecode(img); return t or None
def wechat(img):
    r, _ = wc.detectAndDecode(img); return r[0] if r else None
def zx(img):
    r = zxingcpp.read_barcodes(img, formats=zxingcpp.BarcodeFormat.QRCode); return r[0].text if r else None
def zbar(img):
    g = cv2.cvtColor(img, cv2.COLOR_BGR2GRAY)
    r = pyzbar.decode(g, symbols=[pyzbar.ZBarSymbol.QRCODE]); return r[0].data.decode() if r else None

decs = {"py_zxingcpp": zx, "py_zbar": zbar, "py_opencv": opencv}
if wc: decs["py_wechat"] = wechat
out, times = {}, {}
for name, fn in decs.items():
    t0 = time.perf_counter(); res = {}
    for f in files:
        img = cv2.imread(str(f))
        try: res[f.name] = fn(img)
        except Exception: res[f.name] = None
    times[name] = (time.perf_counter() - t0) / len(files) * 1000
    out[name] = res
json.dump({"results": out, "ms_per_image": times}, sys.stdout)
