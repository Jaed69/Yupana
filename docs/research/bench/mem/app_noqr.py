from fastapi import FastAPI
import anthropic, lxml.etree, sqlalchemy, psycopg, pydantic, httpx  # typical api/worker deps
app = FastAPI()
client = anthropic.Anthropic(api_key="x")
@app.get("/")
def root(): return {"ok": True}
