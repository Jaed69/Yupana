# Capa de ingesta multicanal e integración SUNAT/SIRE para una plataforma de facturas de compra (Perú, 2026)

Notas de investigación al 25/10/2026. Las fuentes primarias (Meta, Telegram, SUNAT) se leyeron directamente; donde la fuente es secundaria o el dato viene de conocimiento previo no verificado en esta sesión, se indica.

## 1. WhatsApp Business Platform (Cloud API): precios, recepción de medios, requisitos, vías no oficiales

### Takeaway
Desde el 1/7/2025 Meta cobra por mensaje entregado (no por conversación), y solo los mensajes de la empresa al usuario. **Cambio clave para este producto:** los mensajes de servicio (respuestas no plantilla dentro de la ventana de 24 h que abre el usuario) eran gratis desde el 1/11/2024, pero **desde el 1/10/2026 se cobran**, con 1.000 gratis al mes por número. Recibir mensajes entrantes (las facturas que manda el usuario) no se cobra. Lo que cuesta es responderle.

### Cited Findings
- Modelo por mensaje desde el 1/7/2025. Solo se cobran los mensajes empresa→usuario y solo si se entregan. La tarifa depende de la categoría y del código de país del destinatario — [Meta: Pricing on the WhatsApp Business Platform](https://developers.facebook.com/docs/whatsapp/pricing/)
- Categorías. Plantillas, que se pueden enviar en cualquier momento: Marketing, Utility y Authentication. Sin plantilla, solo dentro de la ventana de atención de 24 h: Service, y Meta Business Agent desde el 1/7/2026 — [Meta pricing](https://developers.facebook.com/docs/whatsapp/pricing/)
- Service fue gratis desde el 1/11/2024. **Desde el 1/10/2026 se cobra**, con 1.000 mensajes de servicio gratis al mes por número de teléfono, sin acumular. Sin método de pago registrado, los que pasen del tramo gratis no se entregan (Meta pedía registrarlo antes del 30/9/2026). Gobiernos y ONG elegibles siguen gratis hasta el 31/12/2027 — [Meta pricing](https://developers.facebook.com/docs/whatsapp/pricing/)
- Las plantillas utility enviadas dentro de una ventana abierta fueron gratis del 1/7/2025 al 30/9/2026. **Desde el 1/10/2026 se cobran** — [Meta pricing](https://developers.facebook.com/docs/whatsapp/pricing/)
- Perú: las tarifas utility y authentication suben desde el 1/10/2026. La moneda PEN existe desde el 1/4/2026 y la moneda de la cuenta no se puede cambiar después de crearla (hay APIs de migración). Meta solo cambia precios el primer día de cada trimestre — [Meta pricing](https://developers.facebook.com/docs/whatsapp/pricing/)
- Free Entry Point: si el usuario llega desde un anuncio Click-to-WhatsApp (app Android/iOS) y la empresa responde dentro de la ventana, se abre una ventana gratis de hasta 7 días para todas las categorías, salvo Meta Business Agent — [Meta pricing](https://developers.facebook.com/docs/whatsapp/pricing/)
- Tarifas concretas para Perú: no pude obtener una cifra verificada. Un PDF alojado por Salesforce muestra Perú con marketing 86,8 y utility/authentication 24,7, y service 0, pero **sin moneda ni fecha**. No es confiable — [Salesforce PDF](https://www.salesforce.com/en-us/wp-content/uploads/sites/4/documents/PDF/WhatsApp-Business-Messaging-1000.pdf)
- Medios aceptados. Imágenes JPEG/PNG de hasta 5 MB (RGB/RGBA de 8 bits). Documentos de hasta 100 MB: PDF, txt, Word, Excel y PowerPoint — [Meta Cloud API Media reference](https://developers.facebook.com/docs/whatsapp/cloud-api/reference/media/)
- Recepción. El webhook de un mensaje con medio trae un media ID y una URL. **La URL vence a los 5 minutos** y la descarga exige el access token. Si da 404, hay que pedir otra URL a partir del ID. **Los media ID recibidos por webhook vencen a los 7 días.** Un archivo del cliente de más de 100 MB genera el error 131052 — [Meta Media reference](https://developers.facebook.com/docs/whatsapp/cloud-api/reference/media/)
- Política vigente desde el 15/1/2026 (fuente secundaria): Meta prohíbe los asistentes de IA de propósito general en la plataforma y permite los bots estructurados para tareas de negocio — [Zylos Research, ene-2026](https://zylos.ai/research/2026-01-26-whatsapp-api-automation)
- Librerías no oficiales (Baileys, whatsapp-web.js): una fuente secundaria con interés comercial dice que violan los ToS y pueden llevar al baneo permanente del número sin aviso — [checkleaked blog](https://whatsapp.checkleaked.cc/blog/whatsapp-cloud-api-vs-unofficial); [Zylos](https://zylos.ai/research/2026-01-26-whatsapp-api-automation)

### Inferences
- Para un flujo de recepción de facturas, el costo variable lo pone la respuesta: un acuse de recibo o una pregunta de aclaración cuenta como mensaje de servicio. Con 1.000 gratis al mes por número, un MVP con pocos cientos de usuarios activos probablemente quede dentro del tramo gratis. A escala, cada acuse se cobra. Conviene agrupar respuestas (un resumen en vez de un mensaje por factura) y llevar las notificaciones proactivas (vencimientos, resúmenes) a email o Telegram.
- Los medios hay que descargarlos de inmediato al recibir el webhook (la URL dura 5 minutos) y guardarlos en almacenamiento propio. No se puede depender de Meta para recuperarlos después (7 días).
- El límite de 5 MB para imágenes rara vez molesta, porque WhatsApp comprime las fotos. Si se manda la foto "como documento" llega sin comprimir (hasta 100 MB) y es mejor para el OCR. Vale la pena indicárselo al usuario en el onboarding.
- La regla de "sin asistentes de IA de propósito general" pesa: el bot tiene que presentarse como un servicio acotado de recepción y gestión de comprobantes, no como un chat abierto con un LLM.
- Baileys o whatsapp-web.js no son aptos para producción en un SaaS que maneja datos tributarios. Riesgo de baneo del número, sin SLA, y además se rompe la cadena de confianza con el cliente.

### Gaps
- Tarifa exacta en USD/PEN por mensaje para Perú (marketing, utility, authentication, service) desde el 1/10/2026. Hay que descargarla del rate card oficial (CSV/PDF) en la página de pricing.
- Requisitos de onboarding: verificación de Meta Business, límites de mensajería por tier (250/1K/10K/100K…), registro de display name, y diferencias de costo entre BSP (Twilio, 360dialog, Gupshup…) y Cloud API directa. No los verifiqué en esta sesión. Por conocimiento previo: la Cloud API directa no cobra cuota de plataforma, solo las tarifas de Meta, y los BSP suman markup o una cuota mensual. Confirmar.
- No encontré casos documentados de baneos masivos por Baileys ni el texto exacto de los ToS de WhatsApp sobre clientes no autorizados.

## 2. Telegram Bot API

### Takeaway
Es gratis y el límite de descarga lo pone `getFile`: **20 MB por archivo** en la API cloud. Con un Local Bot API Server, la descarga no tiene límite y la subida llega a 2000 MB. Las fotos llegan como un array de `PhotoSize` y los PDF como `Document`.

### Cited Findings
- Versión actual: Bot API 10.3 (24/8/2026) — [Telegram Bot API](https://core.telegram.org/bots/api)
- `getFile`: "For the moment, bots can download files of up to 20MB in size". El enlace `https://api.telegram.org/file/bot<token>/<file_path>` vale "at least 1 hour" — [Telegram Bot API – getFile](https://core.telegram.org/bots/api#getfile)
- Envío desde el bot: hasta 50 MB por multipart (10 MB para fotos). Por URL: 5 MB las fotos y 20 MB el resto — [Telegram Bot API](https://core.telegram.org/bots/api#sending-files)
- `Message.photo` es un array de `PhotoSize`, uno por tamaño disponible (`file_id`, `file_unique_id`, `width`, `height`, `file_size`). `Message.document` trae `file_id`, `file_name`, `mime_type` y `file_size` — [Telegram Bot API](https://core.telegram.org/bots/api)
- Local Bot API Server: descargas sin límite, subidas de hasta 2000 MB, y `getFile` devuelve una ruta local absoluta — [Telegram Bot API](https://core.telegram.org/bots/api)

### Inferences
- Hay que tomar el `PhotoSize` más grande del array para el OCR. Las fotos de Telegram vienen comprimidas, así que conviene pedirle al usuario que las mande "como archivo" (llegan como `document`) cuando la calidad importa.
- Una factura en PDF o una foto casi nunca pasa de 20 MB, así que la API cloud alcanza. No hace falta un servidor local.
- En Perú WhatsApp domina y Telegram es secundario. Telegram sirve como canal de bajo costo para notificaciones proactivas, que en WhatsApp sí se cobran.

### Gaps
- No encontré una página oficial que diga explícitamente "la Bot API es gratuita". En la práctica no tiene tarifas publicadas, pero no lo verifiqué con una cita textual.

## 3. Ingesta por email (XML + PDF)

### Takeaway
No encontré fuentes primarias en esta sesión. Lo que sigue es inferencia y conocimiento previo.

### Cited Findings
- (Ninguno verificado en esta sesión.)

### Inferences
- Patrón habitual (conocimiento previo, no verificado): el proveedor emisor o su OSE/PSE envía al adquirente un correo con el XML UBL firmado y la representación impresa en PDF, a menudo en un ZIP. La opción más simple es una **dirección de reenvío por cliente** (p. ej. `ruc20xxxx@inbox.plataforma.pe`) sobre un servicio de inbound parsing (SES inbound, SendGrid/Mailgun/Postmark inbound), y procesar los adjuntos `.xml`, `.zip` y `.pdf`. La segunda opción es conectar Gmail API o Microsoft Graph con OAuth y filtros por adjuntos XML. Da mejor cobertura histórica, pero la verificación de scopes sensibles de Google (`gmail.readonly`) requiere una evaluación de seguridad que cuesta tiempo y dinero. IMAP con contraseña de aplicación es la opción más frágil y la menos segura.
- Cuando hay XML, la extracción por OCR o LLM sobra: el XML es la fuente de verdad, y el PDF solo se guarda como respaldo visual.

### Gaps
- Requisitos y costo actuales de la verificación CASA de Google para `gmail.readonly` en 2026. Precios de inbound parsing de cada proveedor. Si SUNAT obliga a enviar el XML al adquirente por email (en la práctica el emisor pone el XML a disposición, y el canal varía).

## 4. Código QR de SUNAT y comprobantes físicos

### Takeaway
El QR de la representación impresa está definido en el **Anexo N.° 6 "Aspectos técnicos", numeral 6.4, en la versión de la RS 113-2018/SUNAT (Anexo A)**. Separa los campos con pipe `|` y lleva 10 campos en orden fijo. Su obligatoriedad se postergó varias veces (RS 340-2017 y RS 309-2018) y quedó fijada desde el 1/1/2019. El QR no trae razón social, base imponible ni ítems. Sirve para **leer la clave de la consulta de validez** (RUC, tipo, serie, número, fecha, total), no para reconstruir la factura.

### Cited Findings
- Especificación: simbología QR Code 2005 (ISO/IEC 18004:2006), sin Micro QR, nivel de corrección de errores **Q**, módulo mínimo de 0,190 mm y **codificación UTF-8** — [SUNAT, RS 113-2018, Anexo A (Anexo N.° 6, 6.4)](https://www.sunat.gob.pe/legislacion/superin/2018/anexoA-113-2018.pdf)
- Contenido y orden (texto literal): "RUC | TIPO DE DOCUMENTO | SERIE | NUMERO | MTO TOTAL IGV | MTO TOTAL DEL COMPROBANTE | FECHA DE EMISION | TIPO DE DOCUMENTO ADQUIRENTE | NUMERO DE DOCUMENTO ADQUIRENTE | VALOR RESUMEN", "siendo el separador de campo el carácter pipe ('|')". Los datos van "con el mismo formato empleado en el comprobante" y "en la medida que exista" (un campo sin dato queda vacío). El valor resumen es el `<ds:DigestValue>` en base64 del XML, y "podrá consignarse en la representación impresa, fuera del código QR" — [SUNAT, RS 113-2018, Anexo A](https://www.sunat.gob.pe/legislacion/superin/2018/anexoA-113-2018.pdf)
- Impresión: en la parte inferior de la representación impresa, como máximo 6×6 cm, con una zona de silencio de 1 mm, en negro. Aparte existe el código de barras PDF417 (máx. 2×6 cm) — [SUNAT, RS 113-2018, Anexo A](https://www.sunat.gob.pe/legislacion/superin/2018/anexoA-113-2018.pdf); [SUNAT, RS 357-2015, Anexo C](https://www.sunat.gob.pe/legislacion/superin/2015/anexoC-357-2015.pdf)
- Cronología: la RS 340-2017/SUNAT prorrogó la obligación del QR (fijada originalmente desde el 1/1/2018) y la RS 309-2018 la postergó otra vez. Un blog tributario resume la "flexibilización" de 2018: el valor resumen puede ir fuera del QR — [SUNAT RS 340-2017](https://www.sunat.gob.pe/legislacion/superin/2017/340-2017.pdf); [SUNAT RS 309-2018](https://www.sunat.gob.pe/legislacion/superin/2018/309-2018.pdf); [Escenario Tributario (secundaria)](https://escenariotributario.blogspot.com/2018/05/flexibilizacion-en-codigo-qr-de-uso.html)
- Comprobantes físicos: SUNAT mantiene el trámite de autorización de impresión (Formulario Virtual 816, vía SOL, con imprenta autorizada) y una consulta de validez de comprobantes físicos autorizados — [gob.pe: Obtener autorización para imprimir comprobantes de pago](https://www.gob.pe/390-obtener-autorizacion-para-imprimir-comprobantes-de-pago)
- La API de Consulta Integrada devuelve `estadoCp = "4" : "NO AUTORIZADO" (no autorizado por imprenta)`. Eso indica que también valida comprobantes físicos contra las autorizaciones de imprenta — [SUNAT, Manual Consulta Integrada por Servicio Web v2.0](https://cpe.sunat.gob.pe/sites/default/files/inline-files/Manual-de-Consulta-Integrada-de-Comprobante-de-Pago-por-ServicioWEB_v2_0.pdf)
- La emisión física "en contingencia" está permitida a emisores electrónicos que no puedan emitir de forma electrónica por causas no imputables a ellos, con topes (10% del promedio mensual o 100 formatos por tipo y establecimiento). Fuente secundaria basada en una norma de 2018, posiblemente modificada — [resumen de búsqueda; lpderecho.pe](https://lpderecho.pe/sunat-modifica-normas-emision-electronica/)

### Inferences
- Parser robusto: hacer `split('|')` y aceptar 9 o 10 campos (el valor resumen puede faltar), con espacios alrededor del pipe y campos vacíos (por ejemplo, IGV vacío en operaciones exoneradas, o adquirente vacío en boletas). La fecha suele venir como `YYYY-MM-DD`, que es el formato del XML. La consulta API, en cambio, pide `dd/mm/yyyy`, así que hay que convertirla. En la práctica hay emisores que no cumplen el formato: QR con una URL de su portal, separadores distintos o importes formateados. Tratar el QR como una pista y validarlo siempre con la API de SUNAT.
- Flujo para fotos: (1) decodificar el QR, que da precisión casi perfecta en los campos clave; (2) si no se puede, usar OCR/LLM; (3) consultar validez en SUNAT; (4) cruzar con la propuesta del RCE en SIRE para recuperar los datos oficiales.
- Los comprobantes físicos siguen existiendo (contingencia, contribuyentes no obligados o rezagados, recibos de imprenta), pero son minoritarios. La API los valida (códigos de estado 0 y 4), aunque no hay XML. El único dato estructurado será lo que se extraiga de la foto, más lo que aparezca en el RCE de SIRE.

### Gaps
- No verifiqué si una RS posterior a 2018 (por ejemplo RS 114-2019 o la consolidación de anexos de 2022) cambió el orden de campos del QR. El texto de 113-2018 es el más reciente que leí.
- No hay estadísticas de cuántos emisores cumplen el formato exacto del QR.
- No verifiqué la normativa vigente en 2026 sobre la obligatoriedad residual de comprobantes físicos.

## 5. APIs SUNAT: Consulta Integrada, SIRE (RCE) y modelo de credenciales

### Takeaway
Hay dos APIs con modelos de credenciales distintos. **La Consulta Integrada** usa OAuth `client_credentials` con un client_id/secret generado en SOL. El RUC del path es el del consultante y se puede consultar el comprobante de cualquier emisor. Una sola credencial de la plataforma podría validar comprobantes de todos sus clientes. **La API SIRE**, en cambio, usa un grant tipo `password` que exige **RUC + usuario SOL + Clave SOL del propio contribuyente**, además de un client_id/secret generado por él. Para descargar la propuesta del RCE de cada cliente, la plataforma necesita guardar su Clave SOL, y eso es lo más sensible del diseño.

### Cited Findings
- Consulta Integrada. Token: `POST https://api-seguridad.sunat.gob.pe/v1/clientesextranet/{client_id}/oauth2/token/` con `grant_type=client_credentials` y `scope=https://api.sunat.gob.pe/v1/contribuyente/contribuyentes` ("valores fijos"). Devuelve `access_token` y `token_type`. Las credenciales se generan en SOL, en "Credenciales de API SUNAT" — [SUNAT, Manual Consulta Integrada v2.0](https://cpe.sunat.gob.pe/sites/default/files/inline-files/Manual-de-Consulta-Integrada-de-Comprobante-de-Pago-por-ServicioWEB_v2_0.pdf)
- Consulta: `POST https://api.sunat.gob.pe/v1/contribuyente/contribuyentes/{RUC}/validarcomprobante` con `Authorization: Bearer <token>`. El body JSON lleva `numRuc` (RUC del emisor, 11), el tipo de comprobante (`codComp`), `numeroSerie` (an4), `numero` (hasta 8), `fechaEmision` (dd/mm/yyyy) y el monto total ("solo para" ciertos tipos). La respuesta trae `success`, `data.estadoCp`, `data.estadoRuc`, un campo de condición de domicilio y `data.Observaciones`. Nota: el texto extraído del PDF salió parcial, porque parte de las tablas son imágenes; los nombres `codComp`, `monto` y `condDomiRuc` hay que confirmarlos visualmente en el manual o en Greenter — [SUNAT, Manual Consulta Integrada v2.0](https://cpe.sunat.gob.pe/sites/default/files/inline-files/Manual-de-Consulta-Integrada-de-Comprobante-de-Pago-por-ServicioWEB_v2_0.pdf)
- Códigos. `estadoCp`: 0 NO EXISTE, 1 ACEPTADO, 2 ANULADO, 4 NO AUTORIZADO (imprenta). `estadoRuc`: 00 ACTIVO, 01 BAJA PROVISIONAL, 02 BAJA PROV. POR OFICIO, 03 SUSPENSIÓN TEMPORAL, 10 BAJA DEFINITIVA, 11 BAJA DE OFICIO, 22 INHABILITADO-VENT.UNICA. Condición de domicilio: 00 HABIDO, 09 PENDIENTE, 11 POR VERIFICAR, 12 NO HABIDO, 20 NO HALLADO. Tipos admitidos (`codComp`): 01 Factura, 03 Boleta, 07 NC, 08 ND, R1 Recibo por Honorarios, R7 NC de recibos — [SUNAT, Manual Consulta Integrada v2.0](https://cpe.sunat.gob.pe/sites/default/files/inline-files/Manual-de-Consulta-Integrada-de-Comprobante-de-Pago-por-ServicioWEB_v2_0.pdf)
- Librería de referencia: Greenter `consulta-cpe` (PHP) implementa este flujo con `client_credentials` — [Packagist greenter/consulta-cpe](https://packagist.org/packages/greenter/consulta-cpe)
- SIRE Compras (Manual de servicios Web API SIRE Compras v28). Token: `https://api-seguridad.sunat.gob.pe/v1/clientessol/{client_id}/oauth2/token/` con `grant_type=password`, `scope=https://api-sire.sunat.gob.pe`, `client_id`, `client_secret`, `username = {RUC}{USUARIO}` y `password = {CLAVESOL}` ("Clave sol del generador"). Las credenciales se generan en SOL eligiendo "MIGE RCE y RVIE – SIRE". Si se piden credenciales nuevas, las anteriores se inhabilitan — [SUNAT, Manual API SIRE Compras v28](https://cpe.sunat.gob.pe/sites/default/files/inline-files/Manual%20de%20servicios%20Web%20Api%20-%20SIRE_Compras%20v28.pdf); [cpe.sunat.gob.pe SIRE](https://cpe.sunat.gob.pe/node/158)
- Descargar propuesta RCE (servicio 5.34): `GET https://api-sire.sunat.gob.pe/v1/contribuyente/migeigv/libros/rce/propuesta/web/propuesta/{perTributario}/exportacioncomprobantepropuesta?codTipoArchivo={0 txt|1 csv}&codOrigenEnvio=…&fecEmisionIni=…&fecEmisionFin=…&codTipoCDP=…&numSerieCDP=…&numCDP=…&numDocAdquiriente=…&mtoDesde=…&mtoHasta=…`. Es asíncrono: genera un **ticket**, y con "consultar estado del ticket" (5.31) se obtiene el nombre del archivo cuando el estado es "Terminado". Luego se descarga el archivo, comprimido y particionado — [SUNAT, Manual API SIRE Compras v28](https://cpe.sunat.gob.pe/sites/default/files/inline-files/Manual%20de%20servicios%20Web%20Api%20-%20SIRE_Compras%20v28.pdf)
- El manual advierte: "los servicios del API SIRE no deben ser consumidos desde un cliente Web". Los servicios con carga TUS (resumable uploads) deben hacerse en Java (Anexo 7.4) — [SUNAT, Manual API SIRE Compras v28](https://cpe.sunat.gob.pe/sites/default/files/inline-files/Manual%20de%20servicios%20Web%20Api%20-%20SIRE_Compras%20v28.pdf)
- La obligatoriedad de SIRE sigue ampliándose: desde octubre de 2026 ciertos Principales Contribuyentes deben llevar RCE/RVIE en SIRE (fuente secundaria, solo el titular) — [Nubefact blog](https://blog.nubefact.com/noticias-tributarias)

### Inferences
- **La validación de cada factura** puede hacerse con una sola credencial client_credentials de la plataforma (su propio RUC en el path). El body identifica al emisor por separado (`numRuc`) y no pide el RUC del adquirente, lo que sugiere que se puede consultar cualquier comprobante. Esto es una inferencia y hay que confirmarlo en pruebas. Así se evita pedir credenciales a cada cliente para validar.
- **La sincronización con el RCE** (traer la lista oficial de todas las facturas de compra del periodo, que es el "inbox" más completo y no depende de que el proveedor mande nada) sí requiere la Clave SOL de cada cliente. Lo recomendable es pedirle al cliente que cree en SOL un **usuario secundario** con perfil limitado a SIRE/consultas, no el usuario principal. Hay que guardar la clave cifrada con KMS y una clave por tenant (envelope encryption), con acceso solo desde el worker backend, rotación, auditoría y consentimiento explícito. Una filtración expondría el acceso tributario de todos los clientes. Es el activo más crítico del sistema.
- La propuesta del RCE trae los comprobantes con montos, el RUC del emisor y la serie/número, pero **no el detalle de ítems**. Para categorizar por descripción de ítem hace falta el XML (por email) o la extracción de la imagen/PDF.
- Hay que diseñar para operación asíncrona (tickets, polling) y procesamiento por lotes nocturno por RUC.

### Gaps
- **Rate limits** de ambas APIs: los manuales leídos no los documentan. Hay que medirlos empíricamente y prever reintentos.
- Duración del token (`expires_in`) en ambas APIs: no lo extraje del texto.
- **Descarga del XML de facturas recibidas:** no encontré ninguna API oficial de SUNAT que permita al adquirente bajar el XML de sus comprobantes recibidos. Existe la consulta web en SOL ("Consulta de comprobantes de pago" de recibidos), pero no la verifiqué como API en esta sesión. Queda como gap.
- No hay un mecanismo formal de "delegación de terceros" (OAuth por contribuyente hacia una plataforma) en SIRE. Solo están el usuario secundario SOL y las credenciales propias del contribuyente. No encontré documentación de otra cosa.

## 6. Extracción de datos: OCR vs LLMs de visión, benchmark peruano, UBL 2.1

### Takeaway
El benchmark de Kembec (Grupo Anta, 22/9/2026, 40 documentos ficticios) concluye que **leer el texto del PDF con un LLM da 100% de precisión por campo** (Llama 3.3 70B, Mistral Small 3.1 24B) a menos de US$0,50 por 1.000 documentos. **Leer la imagen** queda claramente por debajo: Llama 4 Scout tiene 94,1% de precisión por campo, pero solo el 57,5% de los documentos salen sin errores. Las reglas determinísticas (dígito verificador del RUC, IGV = 18%, sumas que cuadran, formato de serie) detectan muchos errores, no todos.

### Cited Findings
- Modelos (Cloudflare Workers AI, sin OCR aparte). Texto: Llama 3.3 70B, Mistral Small 3.1 24B y gpt-oss-120b. Imagen: Llama 4 Scout 17B y Llama 3.2 11B Vision. Se evaluaron 40 documentos ficticios (RUC con dígito válido, IGV 18%, montos en PEN y USD), 11 campos y 200 llamadas — [Kembec/benchmark-comprobantes-peru](https://github.com/Kembec/benchmark-comprobantes-peru)
- Resultados (precisión por campo / docs sin error / US$ por 1.000 docs / p50): Llama 3.3 70B (texto) 100% / 100% / 0,49 / 2,6 s. Mistral Small 3.1 (texto) 100% / 100% / 0,34 / 5,2 s. gpt-oss-120b (texto) 77,5% / 77,5% / 0,59 / 9,8 s, porque se quedó sin tokens antes de escribir el JSON en 9 de 40. Llama 4 Scout (imagen) 94,1% / 57,5% / 0,47 / 3,4 s. Llama 3.2 11B Vision (imagen) 66,6% / 2,5% / 0,24 / 4,3 s. El costo total de la corrida fue de US$0,0849 — [Kembec benchmark](https://github.com/Kembec/benchmark-comprobantes-peru)
- Las 4 reglas determinísticas marcaron 26 de las 50 lecturas de imagen que tenían JSON válido con al menos un campo mal. Otras 6 no devolvieron JSON válido. No detectaron errores en el monto exonerado, en la razón social ni un RUC del emisor que en realidad era el del receptor (que igual pasa el dígito verificador). Una corrección del 23/9/2026 pasó a contar aparte los JSON inválidos — [Kembec benchmark](https://github.com/Kembec/benchmark-comprobantes-peru)
- Catálogo N.° 25 (código de producto SUNAT): se actualizó con los anexos 25.1 (códigos con partida arancelaria de referencia), 25.2 (bienes sujetos a detracción) y 25.3 (percepción), con referencia a UNSPSC v26_0801 y códigos de 8 dígitos. Los comprobantes de quienes vendan esos bienes sin el código correcto se rechazan. La vigencia anunciada es el 1/8/2026, pero hay una inconsistencia de fechas en la fuente, y otro post del 1/9/2026 dice que algunas reglas se aplazaron al 1/1/2027 (fuente secundaria) — [Nubefact: Nuevos códigos de productos SUNAT](https://blog.nubefact.com/actualizaciones-sunat/nuevos-codigos-de-productos-sunat-evita-el-rechazo-de-tus-comprobantes-desde-agosto-de-2026); [Nubefact noticias](https://blog.nubefact.com/noticias-tributarias)

### Inferences
- El benchmark es pequeño, usa documentos **sintéticos y limpios** y solo modelos abiertos. No incluye fotos reales de papel arrugado ni tickets térmicos, ni modelos frontera (Claude, GPT, Gemini) ni OCR dedicado (Textract, Document AI, PaddleOCR). Sus cifras de imagen son probablemente un piso para modelos pequeños y un techo optimista para fotos reales.
- Jerarquía de fuentes recomendada: **XML UBL** (exacto) > **QR** (campos clave exactos) > **texto del PDF + LLM** (muy preciso) > **imagen + LLM de visión** (necesita validación) > **validación SUNAT + RCE** como verificación final de RUC, serie/número, fecha y total.
- Para extraer datos útiles del XML UBL 2.1 con fines de categorización (conocimiento previo, no verificado en esta sesión contra la guía SUNAT vigente):
  - `cac:InvoiceLine/cac:Item/cbc:Description` (descripción).
  - `cac:Item/cac:CommodityClassification/cbc:ItemClassificationCode` con `listID="UNSPSC"` (código de producto SUNAT del catálogo 25).
  - `cac:SellersItemIdentification/cbc:ID`.
  - `cac:AccountingSupplierParty` (RUC y razón social del emisor).
  - `cac:TaxTotal/cac:TaxSubtotal` con el código de tributo (catálogo 5) y la afectación del IGV (catálogo 7).
  - `cbc:InvoiceTypeCode` (catálogo 1).
  - `cac:LegalMonetaryTotal/cbc:PayableAmount`.
  - `cbc:DocumentCurrencyCode`.
  - `ext:UBLExtensions` con la firma, de donde sale `DigestValue`, que es el mismo valor resumen del QR.
- El código UNSPSC del ítem es opcional para la mayoría de emisores, así que muchas facturas no lo traerán o traerán códigos genéricos. La categorización tiene que apoyarse sobre todo en la descripción y en el giro (CIIU) del proveedor.

### Gaps
- No encontré benchmarks públicos con fotos reales de comprobantes peruanos ni con modelos frontera u OCR comerciales.
- No verifiqué la guía de elaboración XML UBL 2.1 de SUNAT vigente en 2026 para confirmar los XPaths, ni si el código de producto pasó a ser obligatorio para todos.

## 7. Marco legal: Ley 29733 y su nuevo Reglamento (DS 016-2024-JUS)

### Takeaway
El nuevo Reglamento (DS 016-2024-JUS, publicado el 30/11/2024, con 135 artículos) **está vigente desde el 30/3/2025** y reemplaza al de 2013. Para un SaaS que procesa facturas con DNI/RUC de personas naturales, lo relevante es esto: inscribir los bancos de datos en el registro de la ANPD, oficial de datos personales según tamaño y volumen, notificar brechas en 48 h, restricciones a la transferencia internacional (la nube fuera de Perú requiere cláusulas o consentimiento) y multas de hasta 100 UIT. Varios detalles vienen de una fuente secundaria y hay que verificarlos contra el texto oficial.

### Cited Findings
- El DS 016-2024-JUS se publicó en El Peruano el 30/11/2024. Tiene Título Preliminar, 3 Títulos, 135 artículos, 6 DCF y 2 DCT. El texto está en el portal de la ANPD — [gob.pe ANPD: DS 016-2024-JUS](https://www.gob.pe/institucion/anpd/normas-legales/6554453-16-2024-jus)
- Vigencia desde el 30/3/2025. La portabilidad rige desde el 30/9/2025 (secundaria) — [RecordingLaw](https://www.recordinglaw.com/es/world-laws/world-data-privacy-laws/peru-data-privacy-laws/)
- Oficial de Protección de Datos. Es obligatorio en entidades públicas, cuando se tratan grandes volúmenes o datos sensibles, o cuando tratar datos sensibles es la actividad principal. El plazo escalonado va por ventas anuales: más de 2.300 UIT, 30/11/2025; 1.700–2.300 UIT, 30/11/2026; 150–1.700 UIT, 30/11/2027; menos de 150 UIT, 30/11/2028. Se notifica a la ANPD dentro de 15 días (RD 100-2025-JUS-DGTAIPD). Fuente secundaria — [RecordingLaw](https://www.recordinglaw.com/es/world-laws/world-data-privacy-laws/peru-data-privacy-laws/)
- Brechas: notificar a la ANPD y a los titulares en 48 h cuando haya gran volumen, datos sensibles o daño probable. Fuente secundaria — [RecordingLaw](https://www.recordinglaw.com/es/world-laws/world-data-privacy-laws/peru-data-privacy-laws/)
- Registro de bancos de datos: es obligatorio, gratuito y se hace por la plataforma SIPDP con aprobación automática. No inscribirse es una infracción grave (5–50 UIT). Fuente secundaria — [RecordingLaw](https://www.recordinglaw.com/es/world-laws/world-data-privacy-laws/peru-data-privacy-laws/)
- Transferencia internacional: solo a países con nivel adecuado. A 2025 no había lista completa de adecuación. Sin adecuación caben las cláusulas contractuales tipo, normas corporativas vinculantes, la necesidad contractual o el consentimiento expreso que indique país y destinatario, y el responsable debe evaluar el país de destino. Fuente secundaria — [RecordingLaw](https://www.recordinglaw.com/es/world-laws/world-data-privacy-laws/peru-data-privacy-laws/)
- Sanciones: leve 0,5–5 UIT, grave 5–50 UIT y muy grave 50–100 UIT, con tope del 10% de los ingresos del año anterior. UIT 2026 = S/ 5.500 (DS 301-2025-EF, según la fuente). Fuente secundaria — [RecordingLaw](https://www.recordinglaw.com/es/world-laws/world-data-privacy-laws/peru-data-privacy-laws/)

### Inferences
- **Rol de la plataforma:** frente a los datos de terceros que aparecen en las facturas de sus clientes (proveedores persona natural con DNI o RUC 10, el propio microempresario), lo más probable es que la plataforma actúe como **encargado de tratamiento**. El cliente sería el responsable. Hace falta un contrato o DPA en los términos y condiciones. En cambio, para los datos de sus propios usuarios (registro, Clave SOL, número de WhatsApp) es **responsable** y debe inscribir ese banco de datos.
- **Residencia de datos:** la ley no exige localización en Perú, pero alojar en AWS, GCP o Azure fuera del país es una transferencia internacional. Hay que prever cláusulas contractuales, informarlo en la política de privacidad y, si es posible, pedir consentimiento expreso. Los proveedores de LLM (Anthropic, OpenAI, Cloudflare) que reciben imágenes de facturas también son destinatarios internacionales y se declaran igual.
- La Clave SOL no es "dato sensible" en sentido legal, pero su compromiso tendría efectos graves. Cae en la obligación de medidas de seguridad y, en caso de brecha, en la notificación en 48 h.
- Una startup pequeña (ventas por debajo de 150 UIT) tendría plazo hasta el 30/11/2028 para el oficial de datos, salvo que se la considere de "gran volumen".

### Gaps
- No leí el texto oficial del DS 016-2024-JUS. Faltan los números de artículo exactos sobre el encargado de tratamiento (contenido mínimo del contrato, subencargo), la inscripción de bancos de datos de encargados y la notificación de transferencias internacionales a la ANPD. Todo el detalle operativo viene de una fuente secundaria (RecordingLaw) y hay que confirmarlo en [gob.pe/anpd](https://www.gob.pe/institucion/anpd/normas-legales/6554453-16-2024-jus).
- No verifiqué si existe una Directiva de Seguridad de la Información actualizada en 2025-2026, ni qué umbral define "gran volumen".
- Reserva tributaria: no investigué si hay obligaciones adicionales (Código Tributario) al tratar información tributaria de terceros.
