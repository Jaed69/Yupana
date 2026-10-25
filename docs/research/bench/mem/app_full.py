from fastapi import FastAPI
import anthropic, lxml.etree, sqlalchemy, psycopg, pydantic, httpx  # typical api/worker deps
import cv2, zxingcpp, numpy  # QR stack
app = FastAPI()
client = anthropic.Anthropic(api_key="x")
@app.get("/")
def root(): return {"ok": True}
