# Arquitectura de plataforma modular (tipo Odoo) para un equipo de 1–2 personas (2026)

Contexto: SaaS de módulos (captura de facturas, validación SUNAT, categorización de gastos, consultas en lenguaje natural, reportes, bots multicanal) sobre Python/FastAPI + PostgreSQL + Docker Compose en un único VPS. Notas basadas en ~18 búsquedas/lecturas; varias fuentes son secundarias y se marcan como tales.

## 1. Microservicios vs monolito modular para equipos pequeños; criterios para separar un servicio

### Takeaway
Lo que reportan los practicantes coincide: para un equipo diminuto conviene un monolito modular (una sola unidad de despliegue con módulos de límites explícitos). Un servicio se separa solo cuando un dolor concreto y medido lo justifica, nunca de forma preventiva. Hasta Shopify, con su enorme escala, eligió el monolito modular, y los límites entre módulos se validan ejecutando el código, no con métricas estáticas.

### Cited Findings
- Fowler (MonolithFirst, 3 jun 2015) recomienda empezar casi siempre con un monolito, aunque se espere llegar a microservicios. Da dos razones. La primera es la incertidumbre sobre si el producto servirá: "the premium of microservices is a drag you should do without". La segunda es que los límites entre servicios son difíciles de acertar al inicio y refactorizar entre servicios cuesta mucho: "By building a monolith first, you can figure out what the right boundaries are." El propio Fowler admite que la evidencia es escasa y que su consejo es tentativo — [Martin Fowler, MonolithFirst](https://martinfowler.com/bliki/MonolithFirst.html)
- Fowler presenta como alternativas el monolito cuidadosamente modular, extraer servicios del monolito de forma gradual y tratar el monolito como desechable ("Don't be afraid of building a monolith that you will discard") — [Martin Fowler](https://martinfowler.com/bliki/MonolithFirst.html)
- En Shopify Unite 2019, Kirsten Westeinde defendió el monolito como un buen punto de partida. Según ella, se reestructura cuando el mal diseño empieza a frenar la entrega de features (la "design payoff line"). Shopify eligió el monolito modular, no microservicios — [InfoQ, 2019](https://www.infoq.com/news/2019/07/shopify-modular-monolith)
- Shopify reconoce que componentizar no bastó. Las llamadas cruzadas entre componentes continuaron y los modelos ActiveRecord seguían compartidos, así que hicieron falta "muros" (Packwerk) — [Shopify Engineering, Enforcing modularity with Packwerk](https://shopify.engineering/enforcing-modularity-rails-apps-packwerk)
- La retrospectiva de Packwerk (Shopify, 7 feb 2024) deja varias lecciones:
  - Los desarrolladores agrupaban el código por nombre o dominio y no por cómo se ejecuta. Un modelo de fraude metido en "billing" pertenecía en realidad a la base del grafo.
  - Los "todo files" de violaciones crecían con cada feature.
  - Eliminaron los privacy checks (`app/public`) en la v3.0 porque convertían la herramienta en "an API design tool".
  - Recomiendan medir el progreso arrancando el sistema y corriendo los tests con solo ese paquete cargado: "running code, more than any metric, will always be the best indicator of real progress".
  - Lograron aislar un paquete base ("Platform Essentials") con cero dependencias, pero les tomó meses — [Shopify, A Packwerk Retrospective](https://shopify.engineering/a-packwerk-retrospective)
- Packwerk no da pautas sobre qué convertir en paquete; esa decisión queda en manos del equipo — [Babbel](https://www.babbel.com/en/magazine/modularizing-our-rails-monolith-with-packwerk)
- Amazon Prime Video (2023) pasó un servicio de monitoreo de streams de una arquitectura distribuida (Step Functions + S3) a un único proceso sobre EC2/ECS y redujo los costos de infraestructura en más de 90%. Las causas fueron las transiciones de estado por cada segundo de stream, que chocaban con los límites de cuenta y se cobran por transición, y las llamadas tier-1 a S3 para guardar frames temporales. El equipo dijo que el enfoque distribuido "wasn't bringing a lot of benefits" en su caso — [DevClass, mayo 2023](https://www.devclass.com/ci-cd/2023/05/05/reduce-costs-by-90-by-moving-from-microservices-to-monolith-amazon-internal-case-study-raises-eyebrows/1621790). Matiz: el caso trata de una carga concreta (un pipeline de procesamiento de medios), y extrapolarlo como argumento general contra los microservicios es discutido — [DevClass](https://devclass.com/?p=5603). El post original de primevideotech.com hoy redirige a aboutamazon.com, así que no lo pude leer de primera mano — [URL original](https://www.primevideotech.com/video-streaming/scaling-up-the-prime-video-audio-video-monitoring-service-and-reducing-costs-by-90)
- Kamil Grzybek define el monolito solo como un sistema con exactamente una unidad de despliegue. Exige módulos independientes e intercambiables, cada uno con una interfaz definida — [InfoQ, 2020](https://www.infoq.com/news/2020/01/monolith-architectural-drivers/)
- Patrones de comunicación habituales dentro de un monolito modular: llamadas directas a través de interfaces públicas, más un bus de eventos en proceso para el acoplamiento débil (Spring Modulith, MediatR) — [Milan Jovanović](https://milanjovanovic.tech/blog/modular-monolith-communication-patterns.md)
- Una crítica sostiene que eventos de integración + outbox por defecto entre módulos es "un impuesto" cuando la base de datos ya garantiza commits atómicos — [Viascom, Field Guide](https://medium.com/viascom/the-modular-monolith-a-field-guide-36dcf21a477b) (fuente secundaria)

### Inferences
- Criterios concretos para separar un módulo en su propio proceso o contenedor, sintetizados de lo anterior: (a) necesita un perfil de recursos distinto, como OCR o LLM con CPU/GPU o memoria altas, o bots con conexiones largas o webhooks; (b) necesita escalar o desplegarse con otra cadencia; (c) su fallo no debe tumbar la API principal; (d) tiene dependencias pesadas o en conflicto. Ninguno de estos criterios exige otra base de datos ni otro repositorio. En Docker Compose basta con levantar la misma imagen con otro entrypoint (api, worker, bot).
- Para este producto encaja un monolito modular FastAPI con paquetes `modules/<nombre>/` que expongan una interfaz pública (servicios y eventos) y prohíban importar internals ajenos. Esa regla se puede reforzar en CI con import-linter (dato de entrenamiento, no verificado en esta sesión).
- La lección de Shopify aplicada aquí: organizar los módulos por cómo se ejecutan y dependen (núcleo de identidad/tenant/documentos → extracción → validación SUNAT → categorización → reportes/NL), no solo por el nombre comercial del módulo.

### Gaps
- No encontré la fuente primaria de DHH ("The Majestic Monolith" / "Citadel"). Queda sin citar.
- No leí el texto completo del "Modular Monolith: A Primer" de Grzybek; solo resúmenes secundarios.
- No encontré estudios cuantitativos (más allá de anécdotas) sobre el tamaño de equipo a partir del cual compensan los microservicios.

## 2. Cómo implementa Odoo la modularidad y otros sistemas de módulos/plugins relevantes para FastAPI

### Takeaway
Odoo hace la modularidad con addons declarados en un `__manifest__.py` con `depends`, una base de datos compartida y herencia del ORM que permite a un módulo extender modelos de otro "in place". El sistema es muy potente, pero acopla los módulos de forma implícita. Frappe usa `hooks.py` con eventos por documento y overrides que, si chocan, se resuelven por orden de instalación. En Python, pluggy + entry points ofrece un registro de hooks explícito apto para FastAPI.

### Cited Findings
- En Odoo, el manifest es un `__manifest__.py` con un único diccionario Python. Declara un paquete Python como módulo de Odoo y sus metadatos, incluida la lista `depends` (por ejemplo `'base'`) — [Odoo docs, Module Manifests](https://www.odoo.com/documentation/14.0/es/developer/reference/addons/module.html)
- Herencia en el ORM de Odoo según la referencia oficial (v18):
  - `_inherit` con `_name` definido: "name(s) of parent models to inherit from" (herencia clásica, nuevo modelo).
  - `_inherit` sin `_name`: "name of a single model to extend in-place" (extensión, el mecanismo central para que un addon modifique modelos de otro).
  - `_inherits`: "implements composition-based inheritance: the new model exposes all the fields of the inherited models". Los valores se guardan en el registro enlazado, y si hay campos con el mismo nombre, gana "the last one (in the inherits list order)".
  - `Many2one(delegate=True)` hace lo mismo para un solo campo — [Odoo 18 ORM reference](https://www.odoo.com/documentation/18.0/developer/reference/backend/orm.html)
- La comunidad (OCA) creó la clave `depends_if_installed` para dependencias opcionales, que el core no soporta de forma nativa (al menos en Odoo 9) — [PyPI base_manifest_extension](https://pypi.org/project/odoo9-addon-base-manifest-extension/)
- En el plan gratuito "One App Free" de Odoo no se permiten apps de terceros ni personalizaciones en Python — [ERP Research](https://www.erpresearch.com/pricing/odoo); [Odoo pricing](https://www.odoo.com/pricing)
- Frappe:
  - `hooks.py` es un módulo de variables de nivel superior que el framework lee para conectar la app.
  - `doc_events` permite reaccionar a eventos de ciclo de vida de DocTypes de otras apps; varios apps pueden registrar handlers, que se ejecutan en orden de instalación.
  - `override_doctype_class` (v14+) reemplaza el controlador completo, y si varias apps lo usan, "last-installed app wins, silently breaking other apps".
  - `extend_doctype_class` (v16+) añade mixins y permite que varias apps extiendan el mismo DocType — [Frappe docs hooks](https://docs.frappe.io/framework-git/server-side/hooks); los detalles de override/extend vienen de guías secundarias — [skillselion](https://skillselion.com/skills/openaec-foundation/frappe_claude_skill_package/frappe-syntax-hooks-events)
- En el foro de Frappe aparecen errores reales con overrides: rutas con puntos mal escritas que hacen fallar el override en silencio, y overrides que solo se aplican a documentos ya guardados — [Frappe forum](https://discuss.frappe.io/t/failed-attempt-to-use-override-doctype-class/72211); [Frappe forum](https://discuss.frappe.io/t/override-doctype-class-and-doc-events/80558)
- pluggy es "the crystallized core of plugin management and hook calling for pytest". También lo usan tox, devpi y Kedro.
  - El host define hookspecs y los plugins implementan hookimpls.
  - `pm.hook.<name>(...)` llama a todas las implementaciones y junta los resultados no-`None`.
  - `load_setuptools_entrypoints()` descubre plugins instalados como paquetes vía entry points — [pluggy docs](https://pluggy.readthedocs.io/en/latest/)

### Inferences
- Lección de Odoo/Frappe: la extensión "in place" de modelos ajenos (monkey-patching del ORM, overrides de clase) es lo que da flexibilidad, y también lo que produce los conflictos silenciosos y el orden de instalación como fuente de verdad. Para un equipo de 1–2 personas que controla todos los módulos conviene una versión más estricta:
  - un manifest propio por módulo (nombre, versión, `depends`, permisos, flag de facturación);
  - activación por tenant guardada en una tabla `tenant_modules`;
  - routers FastAPI montados por módulo (`app.include_router`), con una dependencia que devuelva 403/404 si el módulo no está activo para el tenant;
  - extensión mediante eventos y hooks explícitos, no mediante herencia de modelos ajenos.
- Entry points + pluggy solo compensan si habrá módulos de terceros o instalables por separado. Para módulos internos basta un registro explícito en código, que es más simple y más fácil de depurar.
- Cada módulo debería ser dueño de sus tablas (prefijo o schema Postgres por módulo, por ejemplo `invoices.*`, `sunat.*`) y exponer lecturas a otros módulos vía servicio o vista. Así se evita el acoplamiento por tabla compartida que Shopify tuvo que deshacer con ActiveRecord.

### Gaps
- No leí la documentación oficial de Odoo sobre el orden de carga de módulos ni sobre `auto_install` (módulos "puente" que se instalan solos cuando están presentes dos módulos, por ejemplo `sale_stock`). Viene de conocimiento previo y no está verificado aquí.
- No encontré retrospectivas de alta calidad (2023–2026) sobre problemas de upgrade de versión en Odoo causados por la herencia de addons.
- Django apps (`INSTALLED_APPS`, señales) no se investigaron en esta sesión.

## 3. Integración por eventos entre módulos dentro del monolito

### Takeaway
En un único VPS, Postgres puede servir de bus de eventos y de cola. Una tabla outbox escrita en la misma transacción que el cambio de negocio da durabilidad. Workers con `FOR UPDATE SKIP LOCKED` consumen esa tabla, y LISTEN/NOTIFY sirve solo como "despertador" de baja latencia, nunca como fuente de verdad. Así no hace falta Redis ni RabbitMQ al principio.

### Cited Findings
- Comportamiento de NOTIFY según la documentación oficial (PostgreSQL 18):
  - Dentro de una transacción solo se entrega si hay commit: "the notify events are not delivered until and unless the transaction is committed".
  - El payload debe tener menos de 8000 bytes por defecto; para datos más grandes, la documentación sugiere guardarlos en una tabla y enviar la clave.
  - La cola es de 8GB en una instalación estándar, y si se llena, las transacciones con NOTIFY fallan al hacer commit.
  - Los duplicados idénticos dentro de una misma transacción se fusionan.
  - Mensajes de distintas transacciones llegan en orden de commit.
  - Una sesión que hace LISTEN y mantiene una transacción abierta mucho tiempo bloquea la limpieza de la cola — [PostgreSQL docs NOTIFY](https://www.postgresql.org/docs/current/sql-notify.html)
- La documentación entrega las notificaciones a los clientes que hayan ejecutado LISTEN, lo que sugiere que un listener desconectado pierde los mensajes. No es un mecanismo durable (inferencia a partir de la doc, no explícito) — [PostgreSQL docs](https://www.postgresql.org/docs/current/sql-notify.html)
- PgQueuer es una cola de jobs para Python sobre PostgreSQL:
  - lease de jobs con `FOR UPDATE SKIP LOCKED`;
  - LISTEN/NOTIFY para despertar workers, con polling de respaldo "even if notifications are lost";
  - heartbeats, cancelación y reintentos — [PyPI PgQueuer](https://pypi.org/project/PgQueuer/)
- Otras opciones basadas en Postgres: pqrun (asyncpg, pensado para FastAPI, con recuperación de jobs atascados y reintentos) y PGWerk — [PyPI pqrun](https://pypi.org/project/pqrun/); [PyPI pgwerk](https://pypi.org/project/pgwerk/)
- La ventaja principal de usar Postgres como cola es que la mutación de negocio y el trabajo asíncrono comparten transacción, que es la esencia del outbox transaccional — [Level Up, "Postgres as a queue"](https://levelup.gitconnected.com/postgres-as-a-queue-you-probably-dont-need-rabbitmq-yet-7c2812f32bdf) (fuente secundaria)
- La documentación de PostgreSQL (citada de forma secundaria) advierte que SKIP LOCKED da una vista inconsistente de los datos, apta para tablas tipo cola y no para uso general — [Pg::Queue](https://metacpan.org/pod/Pg::Queue)

### Inferences
- Diseño recomendado para este caso:
  1. Tabla `outbox_events(id, tenant_id, type, payload_jsonb, created_at, processed_at)`, insertada en la misma transacción que, por ejemplo, "factura capturada".
  2. Un dispatcher/worker que lee con SKIP LOCKED y llama a los handlers de cada módulo suscrito: validación SUNAT, categorización, notificación al bot.
  3. NOTIFY opcional para reducir latencia.
  4. Handlers idempotentes, deduplicando por `event_id`.
  
  Una sola librería con Postgres (PgQueuer o Procrastinate) cubre cola y outbox sin infraestructura extra.
- Celery/RQ/Dramatiq/arq suman Redis o RabbitMQ como pieza operativa más. Se justifican con alto throughput o si ya existe Redis para caché. Para tareas como OCR, LLM o llamadas a SUNAT con reintentos, una cola sobre Postgres es suficiente a pequeña escala.
- Para reacciones síncronas y baratas dentro del mismo request basta un bus en proceso (un dict de suscriptores). El outbox solo hace falta cuando el efecto es lento, externo o tiene que sobrevivir a caídas.

### Gaps
- No obtuve la documentación de Procrastinate ni comparativas de rendimiento entre PgQueuer y Procrastinate o Celery.
- No encontré benchmarks independientes del throughput de colas sobre Postgres en un VPS pequeño.

## 4. Multi-tenancy en Postgres: shared schema + RLS vs schema-per-tenant vs DB-per-tenant

### Takeaway
A pequeña escala, con muchos tenants pequeños (pymes o personas naturales en Perú), la opción más simple de operar es un schema compartido con `tenant_id` + Row Level Security como defensa en profundidad: una sola migración y un solo pool de conexiones. RLS tiene trampas concretas (owner/superuser bypass, WITH CHECK, fugas de contexto con pooling) que hay que cubrir con tests. Schema-per-tenant o DB-per-tenant se reservan para clientes premium o que exigen aislamiento.

### Cited Findings
- RLS no se aplica a superusuarios ni al dueño de la tabla. Se recomienda `FORCE ROW LEVEL SECURITY` si el rol de la app puede ser dueño, y probar como el rol de runtime con el tenant sin definir esperando cero filas — [Indie Hackers, experiencia práctica](https://www.indiehackers.com/post/postgres-row-level-security-for-a-multi-tenant-saas-what-i-set-up-and-the-bug-that-slipped-past-it-anyway-be5ce43b78)
- Una política con solo USING no impide escribir filas con el `tenant_id` de otro tenant; hace falta `WITH CHECK` — [Indie Hackers](https://www.indiehackers.com/post/postgres-row-level-security-for-a-multi-tenant-saas-what-i-set-up-and-the-bug-that-slipped-past-it-anyway-be5ce43b78)
- Con connection pooling (PgBouncer), las variables de sesión pueden filtrarse entre requests. Se recomienda `SET LOCAL app.current_tenant` (o `set_config(..., true)`) al inicio de la transacción — [AWS Database Blog](https://aws.amazon.com/blogs/database/multi-tenant-data-isolation-with-postgresql-row-level-security); [stackharbor](https://stackharbor.com/en/knowledge-base/pg-row-level-security-multi-tenant/)
- Patrón fail-closed: si la variable de tenant no está definida, la función devuelve NULL y todas las políticas niegan el acceso — [Indie Hackers](https://www.indiehackers.com/post/postgres-row-level-security-for-a-multi-tenant-saas-what-i-set-up-and-the-bug-that-slipped-past-it-anyway-be5ce43b78)
- RLS "is not 'turn it on and forget it'": cambia cómo se autentica la app, la forma de las consultas y la depuración en producción — [Neon guide](https://neon.com/guides/rls-multi-tenant-apps)
- Schema-per-tenant:
  - cada migración se corre N veces;
  - hay que fijar `search_path` en cada request y correr migraciones en el signup;
  - a favor: aislamiento real, restore por tenant y `DROP SCHEMA` para borrar un cliente.
  
  Las fuentes discrepan en el umbral de degradación: "cientos bien, miles con bloat del catálogo", "se rompe pasados ~500 tenants" o "500+ se siente pesado". Son estimaciones anecdóticas, no benchmarks — [devsaran](https://devsaran.com/writing/choosing-a-multi-tenant-data-model); [dev.co](https://dev.co/blog/multi-tenant-saas-postgres-isolation); [codercops](https://blog.codercops.com/blog/multi-tenancy-saas-architecture-patterns-2026)
- DB-per-tenant fragmenta el pooling de conexiones. Hay fuentes que proponen un modelo híbrido: la mayoría en tablas compartidas y los clientes premium en schemas o bases dedicadas — [dev.co](https://dev.co/multi-tenant-saas-postgres-isolation)

### Inferences
- Recomendación para 1–2 devs en un VPS:
  - `tenant_id` NOT NULL en todas las tablas de negocio, con índices compuestos que empiecen por `tenant_id`;
  - RLS con USING + WITH CHECK y FORCE;
  - un rol de app que no sea owner (migraciones con otro rol);
  - `SET LOCAL` por transacción desde una dependencia FastAPI;
  - un test automatizado que verifique que sin tenant se obtienen 0 filas.
  
  Así la puerta queda abierta a un híbrido posterior, moviendo un cliente grande a su propia base.
- Los datos tributarios (facturas, RUC) son sensibles, así que conviene que RLS sea la segunda capa y no la única: el filtro explícito por tenant en el repositorio de la app sigue siendo la primera.

### Gaps
- No consulté directamente la página oficial de PostgreSQL sobre RLS (CREATE POLICY / ALTER TABLE ... FORCE). Los puntos vienen de fuentes de practicantes y de AWS, y conviene confirmarlos allí.
- No encontré mediciones de overhead de RLS en el planner para cargas típicas.

## 5. "Bring your own database" / propiedad de los datos y exportación

### Takeaway
El patrón de mercado más claro es la "sincronización al warehouse del cliente" como feature de pago: Stripe Data Pipeline exporta todos los datos del cliente a Snowflake, Redshift, BigQuery, Databricks o buckets, con refresco periódico. Para un equipo pequeño lo operable es un orden incremental: exportaciones programadas (CSV/Parquet/JSON y Google Sheets), después la sincronización incremental a una base Postgres del cliente, y deja Airbyte o CDC para más adelante, porque Airbyte self-hosted pide recursos y Kubernetes (kind).

### Cited Findings
- Stripe Data Pipeline:
  - conecta directamente con Snowflake, Redshift, BigQuery o Databricks, y también con S3, GCS y Azure Blob;
  - exporta "all your Stripe data" mapeado 1:1 al esquema de la API, más datasets preconstruidos;
  - la carga histórica tarda hasta 6 horas y después refresca cada 3 horas;
  - Stripe afirma que, antes del producto, los usuarios gastaban meses y hasta US$800.000 construyendo integraciones propias para exportar;
  - se cobra por transacción (por ejemplo CHF 0,025 en una página regional) — [Stripe Data Pipeline](https://stripe.com/en-li/data-pipeline/pricing)
- Airbyte self-hosted con abctl:
  - requiere Docker y crea un clúster Kubernetes con kind dentro de un contenedor, instalado vía Helm junto con NGINX Ingress;
  - como piso de quickstart pide 4 CPU, 8 GB de RAM y 20 GB de disco, y tiene un `--low-resource-mode` para menos de 4 CPU que deshabilita el Connector Builder — [Airbyte OSS quickstart](https://docs.airbyte.com/platform/1.8/using-airbyte/getting-started/oss-quickstart); [localtonet, verificado ago 2026](https://localtonet.com/blog/deploy-airbyte-open-source-official-quickstart)
- La licencia de Airbyte self-managed se describe como Elastic License 2.0, que restringe ofrecerlo a terceros como servicio hosted — [getorchestra](https://www.getorchestra.io/knowledge-hub/comparisons/airbyte-open-source-vs-cloud) (fuente secundaria; verificar la licencia exacta de cada componente y conector)

### Inferences
- Escalera de implementación para este producto, en orden de costo operativo:
  1. Export on-demand y programado por tenant (ZIP con CSV/JSON y los XML/PDF originales de las facturas). Es lo mínimo para portabilidad y para dar confianza.
  2. Sincronización a Google Sheets para usuarios no técnicos.
  3. "Sync a tu Postgres/BigQuery": un job que hace upsert incremental por `updated_at` y `tenant_id` hacia credenciales que da el cliente, como módulo de pago al estilo de Stripe.
  4. Una API de lectura documentada con paginación por cursor, que permite a clientes con Airbyte/Fivetran propios usar su conector genérico HTTP.
  
  Correr Airbyte en el mismo VPS compite por recursos (8 GB de RAM como piso) y además puede chocar con su licencia si se ofrece como servicio a terceros.
- Seguridad, inferida de buenas prácticas y no de una fuente específica de esta sesión:
  - las credenciales del destino del cliente se cifran en reposo y nunca se registran en logs;
  - conexión saliente con TLS;
  - el usuario del destino debe tener privilegios mínimos (solo escritura en un schema dedicado);
  - cada job lee con el contexto de su tenant, bajo RLS, para que un bug no mezcle datos;
  - los enlaces de descarga de exports deben ser firmados y caducar;
  - queda registro de auditoría de cada export.
  
  La CDC lógica (replication slots) hacia un tercero expone la base completa y no filtra por tenant de forma natural. No es apta en un schema compartido sin una capa intermedia.
- DB-per-tenant hace trivial "entregar tu base" (un `pg_dump` del cliente), y por eso es un argumento a favor del híbrido para clientes que lo exijan.

### Gaps
- No encontré documentación verificada del conector Postgres CDC de Airbyte ni de los destinos BigQuery/Sheets (requisitos, costos).
- No encontré ejemplos 2023–2026 de SaaS pequeños o medianos que ofrezcan literalmente "bring your own database" (que el cliente aloje la base primaria). Tampoco reverse-ETL (Hightouch/Census) aplicado a este caso.
- No investigué la normativa peruana (Ley 29733 de Protección de Datos Personales) ni las exigencias de conservación de comprobantes de SUNAT que afectan la exportación y el borrado.

## 6. Monetización de productos modulares (breve)

### Takeaway
Odoo monetiza con un gancho gratuito de una sola app y precios por usuario para todas las apps; la personalización y la API externa quedan en el plan más caro. Es un modelo "valioso solo, mejor junto", donde el primer módulo es gratis y el valor de la suite se cobra.

### Cited Findings
- Planes de Odoo:
  - One App Free: una app, usuarios ilimitados, gratis, solo en Odoo Online, sin apps de terceros ni código Python;
  - Standard: todas las apps, US$31,10 por usuario al mes en pago anual (US$38,90 mensual);
  - Custom: añade Studio, multiempresa y External API, y permite Odoo.sh u on-premise; US$46,70 anual o US$58,40 mensual — [Odoo pricing](https://www.odoo.com/pricing); [ERP Research](https://www.erpresearch.com/pricing/odoo)
- Hay cifras distintas en terceros (Capterra lista Custom a US$61). Un post de LinkedIn, no confirmado, habla de una subida de 20% en Custom desde el 24-sep-2026 — [Capterra](https://capterra.com/p/135618/Odoo/pricing/); [LinkedIn](https://dk.linkedin.com/in/alexlyngsoe)
- Stripe cobra la exportación de datos al warehouse como add-on por transacción — [Stripe Data Pipeline](https://stripe.com/en-li/data-pipeline/pricing)

### Inferences
- Para este producto, una opción coherente es dar gratis un módulo gancho (por ejemplo captura de facturas + validación SUNAT básica con límite de documentos) y cobrar por módulo o por bundle con límites de uso. Los consumos costosos (llamadas a LLM para consultas en lenguaje natural, OCR) conviene medirlos por uso. "Sync a tu base/warehouse" puede ser un add-on premium. Técnicamente, esto pide que la activación de módulos por tenant y la medición de uso formen parte del núcleo desde el día 1.

### Gaps
- No investigué el pricing por módulo de otros SaaS de LatAm (Alegra, Siigo, Contífico), que sería la referencia más relevante para el mercado peruano.
