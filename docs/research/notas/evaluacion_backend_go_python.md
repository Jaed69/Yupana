# Evaluación: backend en Go vs Python (FastAPI)

Fecha: 2026-10-25. Resuelve la decisión abierta n.º 1 de `CONTEXT.md` con evidencia medida, no con cifras de blogs.

## Veredicto (revisado 2026-10-25)

**Go para todo el núcleo (`api`, `worker`, `bot`), con zxing-cpp vía cgo para el QR.** Python queda solo como herramienta de desarrollo (evals, scripts), fuera del VPS.

- **Nueva prioridad del fundador:** el VPS tiene 12 GB y va a alojar muchos proyectos más. La eficiencia es una prioridad de producto y de perfil, no una necesidad técnica: con 12 GB, Python también entraría.
- **El riesgo principal de Go desapareció.** zxing-cpp v3.1.1 trae un wrapper oficial para Go (`wrappers/go`, cgo sobre su C API).
  - Con ese wrapper, Go obtiene **la misma precisión que Python: 114/252** (sección 1).
  - En Linux usa **~27–30 MB contra ~130–145 MB** de Python, bajo la misma carga de decodificación (sección 2).
- **Lo que pierde Go no afecta a producción:**
  - Las evals corren en Python o promptfoo en desarrollo.
  - Los PDF y las fotos van a un LLM con visión, según la prioridad de ingesta ya definida.
  - La validación XSD no hace falta, porque la autenticidad la valida SUNAT.
- **Costos aceptados:**
  - El wrapper está marcado como *experimental*. Mitigación: fijar el tag, encapsularlo tras una interfaz `QRDecoder` y dejar gozxing de respaldo.
  - cgo impide un binario estático: la imagen necesita `libZXing.so` y libstdc++, y el build tiene que ser nativo arm64.
  - Se resigna la experiencia previa en FastAPI.

Veredicto anterior (con la RAM del VPS sin conocer): Python para el núcleo. Se descartó por la nueva prioridad de eficiencia y por la paridad de QR medida con cgo.

## 1. Decodificación QR (benchmark propio)

Se generaron 252 fotos sintéticas de una factura en papel con un QR SUNAT real (`RUC|TIPO|SERIE|NUMERO|IGV|TOTAL|FECHA|...`).

- Variables: 3 tamaños de QR (120/160/220 px), corrección de errores M y Q, y 7 niveles de degradación.
- Degradación aplicada: perspectiva, rotación, sombra, desenfoque y movimiento, ruido, reescalado y compresión JPEG, como al reenviar por mensajería.
- Scripts: `docs/research/bench/qr/`.

| Decodificador | L0 | L0.5 | L1.0 | L1.25 | L1.5 | ≥L1.75 | Total | ms/img |
|---|---|---|---|---|---|---|---|---|
| **Python zxing-cpp 3.x** | 36/36 | **35/36** | 17/36 | **17/36** | **9/36** | 0 | **114/252** | 32 |
| Python OpenCV WeChatQRCode | 35/36 | 27/36 | 18/36 | 14/36 | 6/36 | 0 | 100/252 | 195 |
| **Go + zxing-cpp 3.1.1 vía cgo** (Linux, Docker) | 36/36 | **35/36** | 17/36 | **17/36** | **9/36** | 0 | **114/252** | 42 |
| Go gozxing (TRY_HARDER + fallback global) | 32/36 | 18/36 | 16/36 | 11/36 | 4/36 | 0 | 81/252 | 99 |
| Python pyzbar | 36/36 | 12/36 | 6/36 | 0/36 | 0 | 0 | 54/252 | 43 |
| Python OpenCV QRCodeDetector | 25/36 | 13/36 | 5/36 | 3/36 | 0 | 0 | 46/252 | 47 |

- Go con zxing-cpp vía cgo da resultados **idénticos** a Python zxing-cpp: es el mismo motor C++. Build: `docs/research/bench/qr/go-zxingcpp/Dockerfile`, que necesita los submódulos git de zxing-cpp (zint).
- **arm64 verificado:** la imagen compila para `linux/arm64` (con emulación QEMU en Docker Desktop) y da **114/252, resultados idénticos imagen por imagen** a amd64. No se midió la memoria en arm64, porque bajo QEMU el RSS incluye al emulador.
- zxing-cpp decodifica **41 % más** imágenes que gozxing. Con una degradación leve (L0.5) duplica su resultado: 35 contra 18.
- gozxing no decodificó **ninguna** imagen que zxing-cpp no resolviera.
- Ningún decodificador devolvió un texto incorrecto: los errores son siempre "no leído", nunca un dato falso.
- Lo que más falla es el tamaño del QR: los QR de 120–160 px concentran casi todos los fallos. Pedir al usuario que acerque la cámara ayuda más que cambiar de librería.
- **Límite:** son datos sintéticos. Hay que repetir con fotos reales de Basílica Café (`gen.py` se reemplaza por una carpeta de fotos más su `truth.json`).

## 2. Memoria (medición propia)

RSS y USS después de 2.000 requests, un proceso, Windows amd64. El valor es indicativo: en Linux arm64 cambia el valor absoluto, pero no el orden de magnitud. Scripts: `docs/research/bench/mem/`.

| Proceso | RSS | USS |
|---|---|---|
| Go `net/http` + anthropic-sdk-go + pgx (binario estático, 13 MB) | ~17 MB | ~8 MB |
| FastAPI mínimo | ~63 MB | ~41 MB |
| FastAPI + anthropic, SQLAlchemy, psycopg, lxml, httpx | ~141 MB | ~112 MB |
| FastAPI + todo lo anterior + OpenCV/zxing-cpp/numpy | ~156 MB | ~122 MB |

**Linux (Docker amd64), servicio HTTP que decodifica las 252 fotos dos veces (504 decodificaciones):**

| Proceso | VmRSS reposo | VmRSS tras carga | Imagen |
|---|---|---|---|
| Go + zxing-cpp vía cgo (`debian:trixie-slim`) | ~10 MB | **~30 MB** | 135 MB |
| FastAPI + dependencias completas + zxing-cpp/OpenCV (`python:3.13-slim`) | ~151 MB | **~145 MB** | 727 MB |

La medición de Go no incluye el SDK de Anthropic ni pgx; en Windows esas dependencias sumaron solo unos pocos MB.

- La mayor parte del costo en Python la ponen las dependencias (SDK, ORM, lxml), no FastAPI. El stack de QR agrega solo unos 15 MB.
- Una primera corrida de Go marcó 57 MB y no se repitió en 3 corridas más; se trata como un dato atípico.
- Mitigación en Python: un worker por proceso, sin `--workers N`, e importar OpenCV solo en el worker.

## 3. Madurez de librerías

Consultado el 2026-10-25 en GitHub, PyPI y pkg.go.dev.

| Área | Go | Python | Ventaja |
|---|---|---|---|
| QR | gozxing: último tag de 2021, código puro Go | zxing-cpp 3.1.1 (2026-07), wheels aarch64 | Python |
| XML UBL + XSD | `encoding/xml`/etree; el binding de libxml2 está archivado, así que no hay validación XSD | lxml 6.1 con XSD | Python |
| XMLDSig | goxmldsig, pensado para SAML; no está verificado con firmas SUNAT | signxml 5.2 | Python |
| PDF (texto) | ledongthuc/pdf, básico; unipdf es comercial | pypdf / pdfplumber (permisivos). Evitar PyMuPDF (AGPL) | Python |
| SDK LLM | anthropic-sdk-go y openai-go oficiales | oficiales | Empate |
| Evals | promptfoo, que funciona contra HTTP | DeepEval, inspect_ai, promptfoo | Python |
| Postgres / cola | pgx + sqlc + **River** | psycopg 3 + SQLAlchemy 2.1 + procrastinate/pgqueuer | Go |
| pgvector | pgvector-go | pgvector-python | Empate |
| Telegram | go-telegram/bot, telego | aiogram 3 (MIT). Evitar python-telegram-bot (GPL-3.0) | Empate |
| WhatsApp Cloud / SUNAT | sin SDK oficial: cliente HTTP propio | sin SDK oficial (pywa es de la comunidad) | Empate |
| Peru-specific | nada maduro | nada maduro (referencia: greenter, en PHP) | Empate |

## 4. Factores no técnicos

- El fundador ya tiene FastAPI en producción, con 414 tests. Eso es velocidad de entrega desde la Fase 0.
- El cuello de botella es la I/O externa (SUNAT, Meta, LLM), no la CPU. Con unas 200 facturas al día, la ventaja de Go en rendimiento no se nota.

## Pendientes antes de cerrar

1. ~~RAM del VPS~~: 12 GB en total, confirmado por el fundador.
2. Repetir el benchmark de QR con unas 50–100 fotos reales del piloto.
3. Medir en el VPS (arm64 nativo), dentro de un contenedor, la imagen real de `api`/`worker` cuando exista.
4. Confirmar que el build cgo funciona en el runner nativo `ubuntu-24.04-arm` del CI de `luciel-platform`.
