from fastapi import FastAPI
import anthropic, lxml.etree, sqlalchemy, psycopg, pydantic, httpx  # typical api/worker deps
import cv2, zxingcpp, numpy  # QR stack
app = FastAPI()
client = anthropic.Anthropic(api_key="x")
@app.get("/")
def root(): return {"ok": True}

import os
@app.get("/decode")
def decode(f: str):
    img = cv2.imread(os.path.join("/data", f))
    r = zxingcpp.read_barcodes(img, formats=zxingcpp.BarcodeFormat.QRCode)
    return {"text": r[0].text if r else None}

@app.get("/rss")
def rss():
    return [l for l in open("/proc/self/status") if l.startswith("VmRSS")][0].split(":")[1].strip()
