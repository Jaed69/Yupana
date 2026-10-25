FROM python:3.13-slim
RUN pip install -q --no-cache-dir fastapi "uvicorn[standard]" anthropic lxml sqlalchemy "psycopg[binary]" httpx opencv-contrib-python-headless zxing-cpp numpy
WORKDIR /s
COPY app_full.py .
CMD ["uvicorn", "app_full:app", "--host", "0.0.0.0", "--port", "8000"]
