# Panorama competitivo: captura, validación y gestión de compras/gastos con IA para MYPEs peruanas (2026)

Nota general: la investigación se hizo el 25-10-2026. Casi todos los precios salen de páginas de los propios proveedores y pueden estar desactualizados. Varios proveedores peruanos (Rapifac como producto propio, STARSOFT, Contasis, Concar/Real Systems, Perunio, Defontana, Bsale) no se pudieron verificar con fuentes primarias dentro del presupuesto de búsqueda; quedan en Gaps.

## 1. Odoo en Perú: localización, módulos comunitarios y digitalización con IA

### Takeaway
La localización oficial de Odoo 19 cubre plan de cuentas, impuestos, facturación electrónica de ventas y reportes RCE/RVIE/PLE, pero no documenta validación de facturas de proveedor ante SUNAT ni sincronización con la propuesta SIRE de compras. La digitalización OCR/IA de Odoo es genérica (no específica de Perú) y se cobra por crédito (1 documento = 1 crédito, unos 0,10–0,15 € por documento en packs). Lo específico de Perú (SIRE, escaneo QR) lo cubren módulos de terceros.

### Cited Findings
- Odoo 19 Perú: `l10n_pe` (plan de cuentas, impuestos y tipos de documento, "la configuración mínima requerida para que una empresa opere en Perú"); `l10n_pe_edi` (generar y recibir facturas electrónicas según SUNAT); `l10n_pe_reports` (RVIE 14.4, RCE 8.4, RCE 8.5 no domiciliados, PLE 5.1, 5.3, 6.1, 1.1, 1.2); `l10n_pe_edi_stock` (guía de remisión electrónica); `l10n_pe_reports_stock`; `l10n_pe_website_sale`; `l10n_pe_pos` — [Odoo 19 docs Perú](https://www.odoo.com/documentation/19.0/es/applications/finance/fiscal_localizations/peru.html)
- En la parte de la documentación de Odoo 19 Perú que se pudo leer no aparecen SIRE (como tal), validación de facturas de proveedor ante SUNAT ni importación de XML de proveedor. Advertencia: solo se leyó ~2/3 de la página — [Odoo 19 docs Perú](https://www.odoo.com/documentation/19.0/es/applications/finance/fiscal_localizations/peru.html)
- La digitalización de documentos de Odoo usa "OCR and artificial intelligence technologies" para leer y rellenar facturas de proveedor y notas de crédito. Es un servicio IAP con créditos prepagados ("Digitizing one document uses one credit"). Puede ser automática o a demanda. Se cargan documentos manualmente, por alias de correo del diario (PDF/XML; JPEG requiere el alias de la app Documents) o con autopublicación para proveedores elegidos. Busca la orden de compra que coincida — [Odoo docs: Document digitization](https://www.odoo.com/documentation/master/applications/finance/accounting/vendor_bills/invoice_digitization.html)
- Campos que extrae: proveedor, referencia, fecha, referencia de pago (solo formato belga), banco, vencimiento, moneda; líneas: descripción, cantidad, precio unitario, impuestos, subtotal y total. La única excepción de país documentada es que no reconoce impuestos de India; no hay nada específico sobre Perú/LatAm — [Odoo docs](https://www.odoo.com/documentation/master/applications/finance/accounting/vendor_bills/invoice_digitization.html)
- Precios IAP de digitalización: 200 créditos gratis para cuentas nuevas; packs de 100 docs por 15 € (0,15 €/doc), 1.000 por 120 € (0,12), 10.000 por 1.000 € (0,10), 100.000 por 5.000 € (0,05). La misma página también indica "0,05 € por crédito" como base (inconsistencia interna) — [Odoo IAP](https://iap.odoo.com/iap/in-app-services/259)
- Un proveedor externo (ECOSIRE) critica que la digitalización nativa pasa cada página por IAP y consume créditos, con un costo que crece con el volumen; vende un módulo a medida desde 349 USD que no usa créditos (afirmación de marketing) — [ECOSIRE](https://ecosire.com/es/apps/odoo/ai-invoice-bill-ocr)
- Módulos de terceros en la Odoo App Store para Perú: `l10n_pe_invoice_qrcode_scan` (v14) lee el QR de la factura electrónica peruana y crea un borrador de factura de proveedor; si el proveedor no existe, lo crea con el RUC — [Odoo Apps QR scan](https://apps.openerp.com/apps/modules/14.0/l10n_pe_invoice_qrcode_scan); `dv_l10n_pe_sunat_sire_08` (v17, SIRE registro 08 = compras) — [Odoo Apps](https://apps.openerp.com/apps/modules/17.0/dv_l10n_pe_sunat_sire_08); `OdooIAScanner` (v19) — [Odoo Apps](https://apps.odoo.com/apps/modules/19.0/OdooIAScanner); `l10n_pe_edi_odoofact` (v13–19) conecta Odoo con el PSE/OSE Nubefact para enviar y validar comprobantes de venta — [Odoo Apps](https://apps.odoo.com/apps/modules/19.0/l10n_pe_edi_odoofact)
- El repositorio OCA/l10n-peru muestra la rama 19.0 con un solo commit y la tabla de módulos del README vacía, es decir, prácticamente sin módulos publicados en esa rama — [GitHub OCA/l10n-peru](https://github.com/OCA/l10n-peru)

### Inferences
- Odoo deja huecos que llenan partners con módulos sueltos: validación de comprobantes ante SUNAT (consulta de validez), conciliación con la propuesta SIRE de compras, condición de proveedor (habido/activo), detracciones y bancarización. Un nuevo actor podría dar todo eso como un módulo integrado.
- La digitalización de Odoo extrae campos genéricos. No hay evidencia de que reconozca el RUC, la serie-número, el tipo de comprobante SUNAT ni el QR peruano de forma nativa. El QR peruano (RUC|tipo|serie|número|IGV|total|fecha…) permite extraer datos de forma determinista y más barata que el OCR.
- La OCA no tiene una comunidad activa para Perú en v19: la localización comunitaria está fragmentada en módulos comerciales.

### Gaps
- Precio de Odoo por usuario/mes en Perú en 2026 (Standard/Custom): no se consultó la página de precios.
- No se verificó con una prueba real si la digitalización de Odoo funciona bien con facturas/boletas peruanas.
- Funciones de IA de Odoo 18/19 más allá de la digitalización (agentes de IA, preguntas en lenguaje natural): no se investigaron.
- Contenido detallado de los módulos de la App Store (precio, mantenimiento): solo se vieron los listados.

## 2. Actores peruanos

### Takeaway
ContaUno (de RAPIFAC S.A.) es el competidor directo más cercano. Tiene una app móvil donde el cliente toma una foto de la factura y la app extrae los datos y la valida con SUNAT. También recibe PDF/Excel de compras por WhatsApp, se integra con SIRE (descarga propuestas y genera asientos), valida comprobantes de forma masiva y revisa la condición del proveedor, desde S/ 59/mes. Pero su orientación es "software para contadores", no habla de IA ni de categorización automática, y no tiene app en iOS. Alegra Perú se centra en facturación de ventas, con SIRE desde S/ 152/mes y sin IA ni WhatsApp anunciados para Perú.

### Cited Findings
**ContaUno (contauno.com)**
- La titular es RAPIFAC S.A. según sus términos y condiciones; los términos dicen que el usuario sigue siendo responsable de revisar los datos importados automáticamente — [ContaUno T&C](https://contauno.com/TerminosCondiciones)
- Captura de compras: app móvil donde el cliente "solo toma una FOTO de la factura o recibo"; extrae los datos y valida la factura con SUNAT; indica si la compra requiere bancarización o detracción; recibe por WhatsApp PDF y Excel de compras digitalizadas y guías escaneadas; importa compras/ventas en masa desde Excel — [ContaUno Características](https://contauno.com/Home/Caracteristicas)
- SIRE: descarga las propuestas de ventas y compras de SUNAT, genera asientos a partir de la propuesta, verifica la validez de las compras y obtiene el tipo de cambio SUNAT — [ContaUno Características](https://contauno.com/Home/Caracteristicas)
- Validador masivo (incluido el archivo SIRE): clasifica cada comprobante como válido, anulado, no existente o rechazado; revisa si el proveedor está activo y habido; exporta a Excel — [ContaUno Características](https://contauno.com/Home/Caracteristicas)
- También ofrece contabilidad (asientos automáticos, liquidación de IGV/Renta/percepciones/retenciones/ITAN, PLE y SIRE, estados financieros), planillas (5ta, ESSALUD, AFP/ONP, CTS, PLAME/AFPNET) y un panel para estudios contables (asignar empresas, controlar declaraciones, ingresos por cliente) — [ContaUno Características](https://contauno.com/Home/Caracteristicas)
- Precios mensuales: Sistema Contable S/ 59 (1 empresa, 1 usuario); Contable + Planillas S/ 139 (hasta 50 empresas, 2 usuarios); VIP S/ 219 (empresas ilimitadas, 4 usuarios, caja/bancos/inventario y App ContaUno gratis el primer año). Esto implica que la app de captura se cobra aparte en los otros planes. App en Google Play; App Store "Próximamente". No menciona "inteligencia artificial" (usa "Sistema Inteligente"). Dice tener "miles de clientes en todo el Perú" — [ContaUno home](https://contauno.com/)

**Alegra Perú**
- Planes mensuales con IGV: Emprendedor S/ 105 (150 comprobantes, sin SIRE); Pyme S/ 152 (300, con SIRE); Pro S/ 223 (700); Plus S/ 353 (1.500); Premium a cotizar. La página de precios no menciona IA, WhatsApp ni un módulo de gastos — [Alegra Perú precios](https://www.alegra.com/peru/precios/)
- Referencia antigua (2021): los planes iban de S/ 39 a S/ 229, lo que muestra una subida de precios — [Infomercado 2021](https://infomercado.pe/jared-vasquez-entregarles-un-sistema-de-facturacion-de-bajo-costo-a-las-empresas-podria-ayudar-a-formalizarlas/)
- Alegra lanzó en Colombia (oct-2025) un gestor de tareas con asistente IA en WhatsApp, "disponible exclusivamente para contadores"; la integración para pymes estaba "en fase final"; no se menciona Perú — [DPL News](https://dplnews.com/?p=292037). Facturación por WhatsApp en beta en México — [MobileTime](https://mobiletime.la/?p=799119)

**Siigo / Bsale**
- Siigo compró la chilena Bsale y opera en Colombia, Ecuador, Uruguay, Perú y México — [Infomercado](https://infomercado.pe/siigo-compra-la-chilena-bsale-y-alista-us100-millones-para-crecer-en-la-region/)
- En Colombia, Siigo permite crear facturas electrónicas con un audio por WhatsApp que interpreta la IA (ventas, no compras) — [Siigo WhatsApp](https://www.siigo.com/facturacion-electronica-whatsapp/)

**Nubefact**
- Es un PSE/OSE para emitir y validar comprobantes de venta; se integra con Odoo. No se encontró evidencia de un módulo de recepción/registro de compras — [Odoo Apps odoofact](https://apps.odoo.com/apps/modules/19.0/l10n_pe_edi_odoofact)

**Otros**
- INVY Perú: POS e inventario con IA para pymes peruanas, catálogo de ventas por WhatsApp; no registra gastos — [Product Hunt](https://www.producthunt.com/p/invy-peru/invy-peru)
- Investigación académica de la UPN (2025): chatbot de WhatsApp que genera facturas electrónicas por voz con PLN, n8n y la API de SUNAT (emisión, no compras) — [UPN](https://research.upn.edu.pe/en/publications/chatbot-con-ia-integrado-a-whatsapp-para-la-generaci%C3%B3n-autom%C3%A1tica/); [LACCEI](https://proceedings.laccei.org/index.php/laccei/article/view/4988)

### Inferences
- ContaUno ya cubre el flujo de foto → extracción → validación SUNAT → SIRE. La diferenciación no puede ser "capturar y validar" en sí, sino: (a) WhatsApp como canal conversacional real (ContaUno usa WhatsApp para recibir archivos, no un bot que registre); (b) IA para categorizar y preguntas en lenguaje natural; (c) diseño para el dueño del negocio y no para el contador; (d) iOS y QR.
- Los proveedores de facturación (Nubefact, Alegra, Siigo) se enfocan en ventas; el lado de compras/gastos está poco atendido entre ellos.

### Gaps
- No se obtuvieron fuentes primarias de Rapifac (como producto propio), STARSOFT, Contasis, Concar/Real Systems, Perunio, Defontana ni de las funciones de Bsale/Siigo para Perú (captura de compras, SIRE, IA, precios).
- Precio de la App ContaUno por separado; no se verificó la calidad del OCR.

## 3. Comparables internacionales y patrones transferibles

### Takeaway
Los comparables globales (Dext) y latinoamericanos (Payana, Zumma, Buk) muestran patrones claros: varios canales de envío, WhatsApp incluido; agentes de IA que leen el ticket y verifican requisitos fiscales; aprobaciones por política; y consultas financieras en lenguaje natural. Dext cuesta unos 25 USD/mes por 250 documentos.

### Cited Findings
- Dext Prepare (EE. UU.): 25,21 USD/mes (302,50 USD/año) por 250 documentos/mes y 5 usuarios — [Dext US pricing](https://www.dext.com/us/pricing/prepare-direct/); en Australia, 33,58 AUD — [Dext AU](https://dext.com/au/pricing/prepare-direct). Las cifras del Reino Unido de terceros no coinciden entre sí (£24–54) — [Marc Andrews 2026](https://marcandrews.com/dext-review-2026-worth-it-for-uk-small-businesses/)
- Dext acepta documentos por arrastrar y soltar, email, WhatsApp, app, Dropbox y feeds — [Capterra](https://www.capterra.com/p/160189/Dext/pricing/)
- Payana (México, jul-2026): agentes de IA en WhatsApp leen imágenes de comprobantes, identifican el tipo de gasto, extraen los datos y verifican requisitos fiscales (CFDI, RFC, razón social, forma de pago); validan contra políticas internas; permiten consultar saldos, cuentas por pagar y estado de facturas, y autorizar pagos en lenguaje natural. Dice tener más de 1.500 empresas en LatAm (dato propio). Más del 75% de las empresas mexicanas usa WhatsApp Business (Meta/BCG) — [IT Sitio](https://www.itsitio.com/mx/distribucion/whatsapp-ya-gestiona-las-finanzas-de-las-empresas-gracias-a-la-inteligencia-artificial/)
- Zumma (México): se envía la foto del ticket por WhatsApp y la plataforma obtiene la factura (CFDI) en más del 95% de los casos; está orientada a viajes y gastos corporativos — [MobileTime](https://mobiletime.la/?p=821577)
- Buk (Chile): describe rendiciones de gastos por WhatsApp Business API, donde la IA extrae proveedor, fecha y monto — [Buk](https://www.buk.cl/novedades/glosario-finanzas/rendicion-de-gastos-por-whatsapp)
- Tailride: recibos por WhatsApp hacia QuickBooks, Xero, Odoo o Business Central (global, en inglés) — [Tailride](https://tailride.so/es/integrations/whatsapp)
- Zoho Books (MX) incluye escaneo automático de recibos en sus planes — [Zoho Books MX](https://www.zoho.com/es-mx/books/pricing/)

### Inferences
- Patrones que se pueden trasladar a Perú: el equivalente de "verificar CFDI/RFC" es validar el comprobante ante SUNAT y el RUC habido/activo; el de "obtener la factura desde el ticket" (Zumma) sería pedir la factura electrónica al proveedor a partir de una boleta o ticket; las políticas y aprobaciones de gasto; las consultas en lenguaje natural sobre cuentas por pagar y gastos.
- Un precio de referencia razonable: unos 25 USD/mes por 250 documentos (Dext). ContaUno, desde S/ 59, ancla el mercado peruano más abajo.

### Gaps
- No se recolectaron precios ni funciones actuales de Expensify, Ramp, Pleo, Rydoo, Hubdoc ni Contpaqi.

## 4. ¿Existe en Perú/LatAm un producto donde el dueño envía facturas por WhatsApp y quedan registradas como compras?

### Takeaway
En Perú no se encontró ningún bot de WhatsApp que registre facturas de compra y las valide ante SUNAT. Lo más cercano es ContaUno, que recibe por WhatsApp PDF/Excel de compras, pero su captura con IA/foto va por la app. En México (Payana, Zumma) y Chile (Buk) ya existe el flujo de gastos por WhatsApp con IA.

### Cited Findings
- ContaUno: "recepción por WhatsApp" de PDF/Excel de compras digitalizadas y guías — [ContaUno](https://contauno.com/Home/Caracteristicas)
- Payana, Zumma y Buk: ver sección 3 — [IT Sitio](https://www.itsitio.com/mx/distribucion/whatsapp-ya-gestiona-las-finanzas-de-las-empresas-gracias-a-la-inteligencia-artificial/); [MobileTime](https://mobiletime.la/?p=821577); [Buk](https://www.buk.cl/novedades/glosario-finanzas/rendicion-de-gastos-por-whatsapp)
- Alegra (Colombia) tiene un asistente de WhatsApp solo para contadores (tareas, no gastos) — [DPL News](https://dplnews.com/?p=292037)
- Gasti: bot de gastos personales por WhatsApp, no para pymes — [A24](https://www.a24.com/trends/la-herramienta-whatsapp-que-simplifica-el-control-tus-gastos-como-usar-esta-ia-cuidar-tu-dinero-n1398678)

### Inferences
- Hay una ventana para ser el primero en Perú con un bot de WhatsApp que registre compras con validación SUNAT. Es plausible que ContaUno, Alegra o Siigo lo añadan pronto, porque ya tienen el canal de WhatsApp en otros países o funciones.

### Gaps
- No se descarta que existan bots pequeños o estudios contables con flujos propios por WhatsApp (n8n, etc.) que no aparecen en los buscadores.

## 5. Señales de tamaño de mercado

### Takeaway
Perú cerró 2025 con 801.134 mypes formales (frente a 980.357 en 2019), una informalidad del 86,5% y solo el 22,9% de las mypes con algún registro de cuentas. La base formal y direccionable es de cientos de miles de RUC, pero cae. SIRE se volvió obligatorio en 2026 para los principales contribuyentes y se extiende gradualmente al resto.

### Cited Findings
- ComexPerú 2025: 801.134 mypes formales (179.223 menos que en 2019); 99,5% de las empresas; ventas de S/ 164.010 millones (13,7% del PBI, frente a 19,5% en 2019); informalidad del 86,5% (86,8% en 2024; 83,8% en 2019); por sector: comercio 2 millones (85,6% informal), servicios 3,3 millones (88,3%), producción 871.079 (84%); solo el 22,9% lleva algún registro de cuentas; el 57% accede a productos financieros formales — [Forbes Perú, sep-2026](https://forbes.pe/economia-y-finanzas/2026-09-17/comexperu-el-pais-cerro-el-ano-pasado-con-179-223-mypes-formales-menos-que-en-2019/)
- Las cifras totales no coinciden: unas 6,1 millones de mypes en total en 2024 según ComexPerú/ENAHO, frente a 2,33 millones de mipymes formales según PRODUCE (2024); las metodologías difieren — [Forbes Perú 2025](https://forbes.pe/economia-y-finanzas/2025-07-16/mypes-representaron-el-997-de-las-empresas-en-peru-y-mas-del-14-del-pbi-nacional-en-2024); [El Búho](https://elbuho.pe/?p=417263)
- SIRE: obligatorio desde enero de 2026 para los PRICOS con ingresos de hasta 2.300 UIT y desde junio de 2026 para los de más de 2.300 UIT; al resto SUNAT lo incorpora gradualmente y le avisa por el buzón SOL. SUNAT genera la propuesta de RCE desde los comprobantes electrónicos y el contribuyente la acepta o ajusta. La propuesta puede incluir compras ajenas u omitir otras, y define el crédito fiscal de IGV — [Modelo.pe](https://modelo.pe/blog/libros-electronicos-sire-rce-rvie-sunat-2026-obligatorio/)
- La R.S. 125-2026/SUNAT aplazó la obligación para los de más de 2.300 UIT hasta el período octubre de 2026 (mes según la consultora; SUNAT confirma la postergación) — [El Peruano](https://elperuano.pe/noticia/299237-sunat-extiende-plazo-del-sire-para-empresas-con-ingresos-superiores-a-2300-uit); [ByB Consultores](https://bybconsultores.pe/?p=11710). Según otra fuente, los no comprendidos siguen con PLE/Portal hasta septiembre de 2026 — [Gestión](https://gestion.pe/economia/sunat-ajusta-cronograma-del-sistema-integrado-de-registros-electronicos-noticia/)
- SUNAT publica una API de servicios web para SIRE Compras (manual v28), lo que permite integrarse programáticamente con la propuesta — [SUNAT manual API SIRE Compras v28](https://cpe.sunat.gob.pe/sites/default/files/inline-files/Manual%20de%20servicios%20Web%20Api%20-%20SIRE_Compras%20v28.pdf)
- SIRE no reemplaza la obligación de conservar los comprobantes — [Noticiero Contable foro](https://noticierocontable.com/foros/tributarios/facturas-de-compras-en-la-propuesta-sire/)

### Inferences
- Que solo el 22,9% de las mypes lleve registros sugiere una adopción baja de software contable. El comprador real suele ser el contador independiente con muchos RUC, como reflejan el panel multiempresa de ContaUno y sus planes "hasta 50 / ilimitadas empresas".
- La propuesta SIRE convierte "detectar compras faltantes o ajenas" en un dolor concreto y mensual: conciliar lo que el dueño envía con la propuesta tiene valor directo sobre el crédito fiscal.

### Gaps
- % de mypes que usan software contable (dato directo), número de contadores colegiados o independientes y cuántos RUC atiende cada uno en promedio: no se encontraron fuentes.
- Situación SIRE para RER/RMT pequeños y NRUS en 2026: no confirmada.

## 6. Brechas y oportunidades de diferenciación

### Takeaway
Las brechas más defendibles son: un bot conversacional de WhatsApp/Telegram para el dueño (no para el contador) que registre compras validadas ante SUNAT; extracción por QR peruano además de OCR; conciliación automática con la propuesta SIRE Compras vía API; categorización con IA y preguntas en lenguaje natural; y un modelo modular tipo Odoo a precio de mype, con acceso compartido dueño-contador.

### Cited Findings
- ContaUno no menciona IA ni categorización; su WhatsApp recibe archivos; no tiene iOS; su posicionamiento es "software para contadores" — [ContaUno](https://contauno.com/Home/Caracteristicas); [ContaUno home](https://contauno.com/)
- Alegra Perú no anuncia IA, WhatsApp ni gastos, y SIRE está solo desde S/ 152 — [Alegra Perú precios](https://www.alegra.com/peru/precios/)
- La digitalización de Odoo es genérica y se cobra por documento; lo peruano se cubre con módulos sueltos de terceros — [Odoo docs](https://www.odoo.com/documentation/master/applications/finance/accounting/vendor_bills/invoice_digitization.html); [Odoo IAP](https://iap.odoo.com/iap/in-app-services/259); [Odoo Apps QR](https://apps.openerp.com/apps/modules/14.0/l10n_pe_invoice_qrcode_scan)
- La propuesta SIRE puede incluir compras ajenas u omitir otras — [Modelo.pe](https://modelo.pe/blog/libros-electronicos-sire-rce-rvie-sunat-2026-obligatorio/)
- En LatAm ya se validó el patrón de WhatsApp + IA + verificación fiscal + consultas en lenguaje natural (Payana, Zumma) — [IT Sitio](https://www.itsitio.com/mx/distribucion/whatsapp-ya-gestiona-las-finanzas-de-las-empresas-gracias-a-la-inteligencia-artificial/); [MobileTime](https://mobiletime.la/?p=821577)

### Inferences
- Oportunidades concretas: (1) un bot de WhatsApp/Telegram con respuesta inmediata ("Factura F001-123 de RUC X válida, S/ 1.180, categoría: insumos, sujeta a detracción"); (2) priorizar el QR para comprobantes electrónicos y usar OCR solo de respaldo, por costo y precisión; (3) alertas de compras que están en SIRE pero el dueño no envió, y de compras enviadas que no están en SIRE; (4) duplicados por RUC + serie + número; (5) alertas de proveedor no habido o con baja, bancarización y detracciones (ContaUno ya lo hace: es lo mínimo esperado); (6) un panel para el contador multi-RUC además del bot para el dueño (modelo de distribución B2B2C vía contadores); (7) gestión de boletas y tickets que no dan crédito fiscal, como gasto pero no como crédito de IGV.
- Riesgos: ContaUno/Rapifac puede añadir un bot e IA rápidamente; la base de mypes formales cae; la disposición a pagar es baja (anclas de S/ 59–105/mes).

### Gaps
- No hay datos de disposición a pagar ni de churn de mypes peruanas por software contable.
- No se verificó si los competidores no revisados (STARSOFT, Contasis, Concar, Perunio, Defontana, Bsale) ya ofrecen captura por IA o bot.
