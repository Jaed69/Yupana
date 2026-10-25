# Capa de IA para plataforma de gastos/facturas: categorización automática y consultas en lenguaje natural (estado a octubre 2026)

## 1. Enfoques de categorización de gastos (reglas, ML clásico, embeddings + kNN, LLM zero/few-shot, híbridos) y cómo lo hacen los productos

### Takeaway
La evidencia apunta a un enfoque híbrido en capas: reglas/memoria por proveedor y kNN sobre el historial etiquetado *del propio tenant* (la categorización es personal: cada empresa tiene su propio plan de cuentas), con un LLM barato zero-shot para cold start y casos nuevos, y corrección del usuario que alimenta la memoria. Los embeddings sirven como mecanismo de "memoria/recuperación", no como clasificador por sí solo; "una base vectorial para categorizar" es, como mucho, un componente menor.

### Cited Findings
- **QuickBooks (Intuit), paper ECML-PKDD 2025:** la categorización se modela por empresa porque "Businesses have unique needs and preferences for how transactions are categorized"; cada *Category* del usuario se agrupa en un *Code* mapeado al código tributario del IRS. — [arXiv 2506.09234](https://arxiv.org/html/2506.09234v1)
- Resultados Top-1 en QuickBooks (15K transacciones de test, 7.5K empresas; entrenamiento ~3M transacciones, 100K categorías, 15K empresas): baseline poblacional zero-shot (Word2Vec + contrastivo, "Shorthair") 36.07%; Txn-Bert zero-shot 43.83–45.52% (Top-5 74–75%); regresión logística por empresa ("Lynx", few-shot) 62.49%; **TopK nearest neighbor sobre historial de la empresa (similitud ≥0.8) 65.80%** (Top-5 78.55%); grafo relacional Rel-Cat 68.67% (Top-5 88.04%). — [arXiv 2506.09234](https://arxiv.org/html/2506.09234v1)
  - Implicancia clave: kNN simple sobre el historial propio ya supera al modelo por empresa en producción; el zero-shot sin historial es mucho peor (36–45%) porque las categorías son idiosincráticas por empresa. En el subconjunto "Historical Unseen" (comercios nunca vistos por esa empresa) kNN cae a 0.16% Top-1 y Rel-Cat a 20.84%. — [arXiv 2506.09234](https://arxiv.org/html/2506.09234v1)
  - El paper no compara contra LLMs generativos (solo encoders tipo Sentence-BERT/MPNet). — [arXiv 2506.09234](https://arxiv.org/html/2506.09234v1)
- **Categorización de ítems de recibos con LLM (arXiv, abril 2026, AWS Bedrock):** 389 ítems, 26–27 categorías fijas. Accuracy zero-shot: Claude 3.7 Sonnet 90.2%, Claude 4 Sonnet 84.8%, Mixtral 8x7B 69.4%, Mistral 7B 59.6%. Con categorías refinadas + reglas en el prompt (zero-shot) sube a 93.3% estricto / 95.6% laxo; **agregar ejemplos few-shot (92.5%) no mejoró y costó más**, por lo que recomiendan zero-shot con reglas. Costo ~US$0.004–0.011 por llamada con Sonnet (421–1,212 tokens de entrada, ~25 de salida). — [arXiv 2604.01615](https://arxiv.org/html/2604.01615v1)
  - Nota: taxonomía fija y pequeña, dataset pequeño; no es comparable directamente al caso "plan de cuentas propio de cada empresa".
- **Mercado Libre (caso práctico, LatAm):** 2021 regex por país + MCC ≈60% accuracy; 2023 GPT-3.5 Turbo con validación humana 80%+; 2024 GPT-4o-mini; 2024–2025 embeddings tipo BERT propios entrenados con etiquetas generadas por LLM ≈90% y ~30% menos costo. Motivo del cambio: latencia y costo de llamar APIs LLM a cientos de millones de transacciones semanales (destilación LLM → modelo pequeño). Fuente promocional, sin metodología detallada. — [ZenML resumen](https://zenml.io/llmops-database/financial-transaction-categorization-at-scale-using-llms-and-custom-embeddings); [Medium MercadoLibre Tech (original)](https://medium.com/mercadolibre-tech/la-nueva-babel-financiera-ense%C3%B1ar-a-la-ia-a-hablar-dinero-en-latinoam%C3%A9rica-4605235e3aac)
- **Ramp:** post de ingeniería (ago 2024) sobre "accounting coding" (asignar transacciones a las categorías del ERP de cada cliente) usando embeddings de transacciones enriquecidas (comercio, MCC, departamento, memo) para recuperación de transacciones similares. — [Ramp Engineering](https://engineering.ramp.com/transaction-embeddings) (no pude obtener el texto completo; resumido vía [búsqueda/ZenML](https://www.zenml.io/llmops-database/using-rag-to-improve-industry-classification-accuracy))
  - Ramp (2025, vía resumen de terceros): agente LLM + RAG para corregir clasificación de comercios; recuperación por embeddings de nombres + fuzzy matching que le da candidatos al LLM; reporta 99% accuracy evaluado por LLM. — [ZenML](https://zenml.io/llmops-database/ai-powered-merchant-classification-correction-agent)
  - Ramp, clasificación de industria NAICS: retrieval por embeddings + re-ranking con LLM; hasta +60% en métricas de recuperación y +5–15% en accuracy final. — [ZenML](https://www.zenml.io/llmops-database/using-rag-to-improve-industry-classification-accuracy)
  - Un VP de Producto de Ramp habría dicho que pasaron de un modelo propio a un LLM con mejores resultados (fuente secundaria). — [Sacra](https://sacra.com/chat/h/b21fdfbb-ef3c-45f3-a645-e3836d1450f0)
- **Taxonomía estándar SUNAT:** el "Código de producto SUNAT" proviene del catálogo N.º 25 (Anexo V, R.S. 340-2017/SUNAT), basado en UNSPSC (v14_0801); también se acepta GS1. — [SUNAT CPE código de producto](https://cpe.sunat.gob.pe/informacion_general/codigoproducto)
  - Se planeó obligatorio desde 2019 ([R.S. 042-2018](https://www.sunat.gob.pe/legislacion/superin/2018/042-2018.pdf)), se postergó/gradualizó ([R.S. 309-2018](https://www.sunat.gob.pe/legislacion/superin/2018/309-2018.pdf)) y según SUNAT la R.S. 133-2019 dispuso que **no es requisito mínimo** de la factura/boleta electrónica. — [SUNAT conceptos generales](https://cpe.sunat.gob.pe/informacion_general/conceptos_generales)
  - Uso obligatorio en casos específicos, p. ej. gastos deducibles de persona natural (tipo de operación 0112) con códigos como 80131501 (arrendamiento) y 84121901. — [Anexo R.S. 123-2022](https://www.sunat.gob.pe/legislacion/superin/2022/anexo-123-2022.pdf)

### Inferencias
- Para una pyme peruana, la señal más fuerte es el **historial del propio tenant** (mismo RUC proveedor + descripción similar → misma cuenta). Una tabla "proveedor/ítem → categoría" con matching exacto/fuzzy probablemente resuelve la mayoría de facturas recurrentes sin IA; kNN por embeddings cubre variaciones de descripción.
- El LLM zero-shot con lista de categorías y reglas en el prompt es el mejor arranque en frío (cliente nuevo sin historial) y para inferir "para qué es" un ítem; luego las correcciones del usuario pasan a la memoria/kNN (efecto: el costo LLM baja con el tiempo). Esto replica el patrón Ramp "retrieval de candidatos + LLM decide".
- UNSPSC/codProdSUNAT, cuando venga en el XML UBL, es una feature útil (prior de alto nivel: "alimentos", "limpieza") pero no puede ser la fuente principal: es opcional, muchos emisores lo llenan genérico o vacío, y no mapea 1:1 a la cuenta PCGE ni a categorías de gestión ("insumos de cocina").
- Conviene separar dos niveles: categoría de gestión editable por el usuario (lo que pregunta el dueño) y mapeo a cuenta PCGE (para contador), mapeo que puede ser regla fija por categoría.
- Mostrar confianza y pedir confirmación cuando kNN/LLM discrepan o la similitud es baja (Top-k sugerencias; Top-5 en QuickBooks es 88% vs 69% Top-1).

### Gaps
- No encontré documentación técnica primaria actual de Brex, Dext, Xero o QuickBooks (producto) sobre su motor de categorización más allá del paper de Intuit; no reporto cómo lo hacen.
- No hallé un benchmark público que compare head-to-head reglas vs kNN vs LLM sobre el mismo dataset de facturas con plan de cuentas por empresa, ni ninguno en español/Perú.
- No verifiqué si hay normativa SUNAT posterior a 2022 que vuelva obligatorio el código de producto.
- No se encontró ningún dataset público de facturas peruanas etiquetadas por categoría o cuenta PCGE.

## 2. ¿Cuándo se justifica una base vectorial dedicada vs pgvector en Postgres vs no usar embeddings?

### Takeaway
A escala pyme (cientos a decenas de miles de facturas por tenant; incluso millones de ítems en total en la plataforma) una base vectorial dedicada no se justifica: pgvector dentro del mismo Postgres (con filtro por tenant y RLS) alcanza de sobra, y para mucha parte de la categorización ni siquiera hacen falta embeddings (matching exacto por RUC proveedor/descripción).

### Cited Findings
- Las guías sitúan el techo práctico de pgvector en "millones" de vectores en una sola instancia, frente a cientos de millones en Qdrant distribuido; una base dedicada gana claramente en la escala de miles de millones (tiempo de construcción de índice y memoria en un nodo). — [ecorpit](https://ecorpit.com/pgvector-vs-dedicated-vector-database-rag-decision-2026/); [encore.dev](https://encore.dev/articles/pgvector-vs-qdrant)
- El punto débil de pgvector es la búsqueda filtrada: con el default `hnsw.ef_search = 40`, un filtro que coincide con el 10% de las filas devuelve ~4 resultados (filtrado posterior al índice). — [The New Stack](https://thenewstack.io/why-pgvector-benchmarks-lie/)
- pgvector indexa vectores `vector` de hasta 2,000 dimensiones (almacena hasta 16,000); para más hay que usar `halfvec`/cuantización. — [The New Stack](https://thenewstack.io/why-pgvector-benchmarks-lie/)
- Benchmarks pgvector vs Qdrant son mayormente de vendors; una comparación independiente en 1.18M vectores encontró diferencias de ~2x en latencia y ~4x en disco/memoria a igual recall. Recomiendan medir con tu propia carga. — [The New Stack](https://thenewstack.io/why-pgvector-benchmarks-lie/); [computingforgeeks](https://computingforgeeks.com/qdrant-weaviate-milvus-pgvector/)

### Inferencias
- Con decenas de miles de ítems por tenant, una búsqueda kNN por tenant puede hacerse incluso por fuerza bruta (sin índice ANN) en milisegundos; el problema de filtrado de HNSW se evita filtrando primero por `tenant_id` (índice B-tree / particionado) y luego ordenando por distancia, o usando índices parciales/particiones por tenant grande.
- Tener vectores y datos transaccionales en el mismo Postgres permite una sola fuente de verdad, transacciones, backups y la misma RLS multi-tenant; una base vectorial separada duplica el problema de aislamiento entre tenants y de sincronización.
- Umbral razonable para reconsiderar: decenas/cientos de millones de vectores globales con búsqueda filtrada a baja latencia p99, o necesidad de búsqueda híbrida avanzada a gran escala. Es una inferencia a partir de las guías, no una cifra medida.

### Gaps
- No encontré benchmarks independientes y rigurosos de pgvector con filtros por tenant a escala multi-tenant SaaS (muchos tenants pequeños).
- No revisé la documentación oficial de pgvector (iterative index scans en versiones recientes, que mitigan el problema de filtrado); conviene verificarla.

## 3. Preguntas en lenguaje natural sobre datos financieros estructurados: RAG vs text-to-SQL vs capa semántica / tool-calling

### Takeaway
RAG (recuperar chunks top-k y que el LLM sume) es la herramienta equivocada para "¿cuánto gasté en X en septiembre vs agosto?": las agregaciones requieren cálculo exacto sobre *todas* las filas, que RAG no garantiza. Lo correcto es que el LLM traduzca la pregunta a una consulta estructurada que ejecuta la base de datos; y para un dominio acotado como gastos, la opción más precisa y segura es una capa semántica / funciones tipadas (tool-calling con métricas predefinidas como `gasto_total(categoría, periodo)`), dejando text-to-SQL libre solo como fallback restringido.

### Cited Findings
- **RAG en agregaciones:** en una charla del AI Engineer World's Fair 2025 (AI21 Labs), un pipeline RAG básico respondió correctamente 5% de preguntas agregadas y un setup de OpenAI Responses 11%; la propuesta fue extraer el corpus a un esquema y responder vía SQL (metodología no detallada). — [AI Engineer](https://ai.engineer/talks/rag-evaluation-is-broken-here-s-why-and-how-to-fix-it)
- Un practicante reporta que los totales por categoría de gasto de un RAG sumaban menos de la mitad del total real; ventanas de contexto más grandes no lo arreglaron (anécdota). — [Towards Data Science](https://towardsdatascience.com/larger-context-windows-dont-fix-rag-so-i-built-a-system-that-does/)
- Anyscale documenta errores de razonamiento numérico/agregación como falla típica de RAG; IBM propone "SQL RAG" (LLM → SQL) porque las bases vectoriales están hechas para similitud, no agregación. — [Anyscale docs](https://docs.anyscale.com/rag/quality-improvement); [IBM](https://www.ibm.com/think/insights/rag-problems-five-ways-to-fix)
- **Benchmarks text-to-SQL:** BIRD (bases reales "sucias"): humano ~92.96%; en sep 2026 el mejor sistema del leaderboard (Huawei DataGallery) reporta 82.39% execution accuracy; Gemini-SQL2 80.04% en el track single-model (sin modelo público). Splits/tracks no son comparables entre sí. — [Technology Magazine](https://technologymagazine.com/news/huawei-datagallery-tops-benchmark-for-enterprise-text2sql-ai); [AI Weekly](https://aiweekly.co/alerts/googles-gemini-sql2-tops-bird-text-to-sql-at-8004); [paper BIRD](https://arxiv.org/abs/2305.03111v2)
- **Spider 2.0** (workflows empresariales reales): o1-preview con agente resolvió ~17.0% (v1) / 21.3% (v2 marzo 2025) de las tareas, vs 91.2% en Spider 1.0 y 73.0% en BIRD; GPT-4o ~10.1%, Claude 3.5 Sonnet ~9.0%. — [arXiv 2411.07763](https://www.alphaxiv.org/abs/2411.07763v2); [hyper.ai leaderboard](https://hyper.ai/ja/sota/tasks/text-to-sql/benchmark/text-to-sql-on-spider-2-0)
- **Capa semántica (paper independiente, abril 2026):** 100 preguntas sobre Contoso Retail en ClickHouse, Claude Opus 4.7, Claude Sonnet 4.6 y GPT-5.4; solo esquema: 45.5–50.5% accuracy; esquema + documento semántico de 4 KB (medidas, convenciones, reglas de desambiguación): 67.7–68.7% (+17–23 pp, p<0.01). Los modelos fueron estadísticamente indistinguibles entre sí → el contexto semántico pesa más que el modelo. — [arXiv 2604.25149](https://arxiv.org/abs/2604.25149)
- **dbt Labs (vendor, abril 2026):** para preguntas cubiertas por una capa semántica bien modelada, accuracy cercana o igual a 100% (p. ej. Sonnet 4.6 90.0%→98.2%; GPT-5.3-Codex 84.1%→100%); la capa semántica responde "no puedo" fuera de lo modelado. Cifras de vendedor. — [dbt blog](https://docs.getdbt.com/blog/semantic-layer-vs-text-to-sql-2026)
- Strategy (vendor) reporta text-to-SQL directo 88.2% global pero 0% en consultas complejas multi-tabla (17 preguntas, un esquema). — [Strategy](https://software.strategy.com/blog/70-80-percent-accuracy-isnt-good-enough-for-enterprise-ai)
- **Seguridad:** "prompt-to-SQL injection" (P2SQL): Pedro et al. (ICSE 2025) evaluaron 7 LLMs y encontraron vulnerabilidades en 5 apps reales LangChain/LlamaIndex; proponen 4 defensas. Trabajo posterior (IEEE S&P 2026) cubre inyección indirecta: texto malicioso sembrado en datos que el asistente luego lee (relevante: descripciones de facturas las escribe un tercero, el proveedor). — [arXiv 2308.01990](https://arxiv.org/abs/2308.01990v2); [ICSE 2025](https://conf.researchr.org/details/icse-2025/icse-2025-research-track/31/Prompt-to-SQL-Injections-in-LLM-Integrated-Web-Applications-Risks-and-Defenses); [INESC-ID S&P 26](https://syssec.dpss.inesc-id.pt/papers/pedro_ieeesp26.pdf)
- Practicantes: reglas en el prompt y validadores regex/AST son controles débiles; el aislamiento multi-tenant debe ser "duro": RLS de la base de datos filtra filas sin importar qué SQL genere el agente, y los datos no permitidos deben estar estructuralmente ausentes. (Fuentes nivel blog.) — [Viblo RLS text-to-SQL](https://viblo.asia/p/row-level-security-cho-text-to-sql-llm-XP4WEwqBL7G); [dotzlaw](https://dotzlaw.com/projects/txttosql-isolation-core/); [dpriver](https://www.dpriver.com/blog/text-to-sql-security-10-risks-before-production-deployment/)

### Inferencias
- Patrón recomendado: el LLM hace *parsing de intención* → llamada a función tipada, p. ej. `comparar_gasto(categoria="insumos de cocina", periodo_a="2026-09", periodo_b="2026-08", moneda="PEN")`; el backend ejecuta SQL parametrizado ya escrito y testeado, con `tenant_id` tomado de la sesión (nunca del LLM). El LLM solo redacta la respuesta con los números devueltos. Esto elimina inyección SQL, garantiza aislamiento y hace los números deterministas y auditables.
- La parte "difusa" de la pregunta ("insumos de cocina") es donde sí ayudan embeddings/LLM: resolver el término del usuario a una o varias categorías/proveedores existentes del tenant (entity linking), mostrando al usuario qué interpretó ("incluí: Abarrotes, Carnes, Verduras").
- El esquema de gastos de una pyme es pequeño y estable (facturas, ítems, proveedores, categorías, fechas, montos, impuestos), así que unas 10–20 métricas/funciones cubren la mayoría de preguntas; text-to-SQL libre como fallback solo sobre vistas de lectura, rol de base de datos read-only, RLS por tenant, límites de tiempo/filas y SQL validado.
- Detalles de dominio que el LLM suele fallar y que deben vivir en la capa semántica: fecha de emisión vs fecha de registro, IGV incluido o no, PEN vs USD (tipo de cambio), notas de crédito que restan, facturas anuladas.

### Gaps
- No obtuve el blog de Uber QueryGPT ni el de Pinterest text-to-SQL (no los busqué por límite de llamadas); no cito sus cifras.
- No hay benchmarks públicos de text-to-SQL en español ni sobre esquemas contables de pymes.
- Las cifras de dbt/Strategy/Atlan son de vendors con interés comercial.

## 4. Dónde sí tiene sentido RAG en este producto

### Takeaway
RAG/búsqueda semántica es útil para texto no estructurado y búsqueda "encuéntrame", no para números: buscar facturas por descripción ("la factura del técnico de la campana"), contratos y condiciones con proveedores, documentos adjuntos, notas, y FAQ normativa (SUNAT/IGV) del asistente.

### Cited Findings
- Las bases vectoriales están diseñadas para búsqueda por similitud, no para agregación; el patrón recomendado es SQL para preguntas numéricas. — [IBM](https://www.ibm.com/think/insights/rag-problems-five-ways-to-fix)
- Ramp usa retrieval por embeddings como paso de búsqueda de candidatos (transacciones similares, comercios candidatos) antes de que un LLM/regla decida, es decir, RAG como "retrieval de contexto", no como motor de cálculo. — [Ramp Engineering](https://engineering.ramp.com/transaction-embeddings); [ZenML](https://zenml.io/llmops-database/ai-powered-merchant-classification-correction-agent)
- Chunk-and-embed falla en tablas financieras; chunking consciente de tablas solo ayuda parcialmente (resumen de un estudio, original no verificado). — [Modelwire](https://themodelwire.com/article/financial-documents-expose-rags-structural-limits-with-tables-01KZD5E2Y5ZATCJK80MX39H41Y)

### Inferencias
- Casos con buen encaje: (1) búsqueda semántica de ítems/facturas; (2) preguntas sobre contratos/términos de proveedores (PDF); (3) resolver términos del usuario a categorías/proveedores (entity linking para la capa de consultas); (4) few-shot dinámico: recuperar ejemplos ya corregidos del tenant para el prompt de categorización.
- Combinación útil: búsqueda híbrida (full-text de Postgres `tsvector` en español + pgvector) dentro del mismo Postgres.
- Los textos de facturas/contratos son de terceros: tratarlos como datos no confiables al pasarlos al LLM (riesgo de inyección indirecta, ver P2SQL arriba).

### Gaps
- No encontré estudios específicos sobre RAG sobre contratos de proveedores de pymes ni métricas de búsqueda semántica en facturas en español.

## 5. Evaluación: cómo armar un eval set para categorización y para consultas NL; métricas

### Takeaway
Ambos componentes se evalúan con conjuntos etiquetados propios y métricas de exactitud determinista: para categorización, accuracy Top-1/Top-k por tenant (con y sin historial) y tasa de corrección del usuario en producción; para consultas, *execution accuracy* (¿el número devuelto coincide con el resultado de un SQL de referencia?) más tasa de abstención correcta.

### Cited Findings
- QuickBooks evalúa con Top-1/Top-2/Top-5, separando zero-shot (sin historial) y few-shot, y con un subconjunto de transacciones de comercios no vistos ("Historical Unseen"); test con las 2 transacciones etiquetadas más recientes por empresa (split temporal). — [arXiv 2506.09234](https://arxiv.org/html/2506.09234v1)
- El estudio de recibos usa accuracy estricta y laxa (laxa acepta categorías razonablemente equivalentes) y balanced accuracy (clases desbalanceadas), más costo por llamada y tokens. — [arXiv 2604.01615](https://arxiv.org/html/2604.01615v1)
- BIRD/Spider usan execution accuracy (EX): comparar el resultado de ejecutar el SQL generado contra el SQL de referencia; BIRD agrega R-VES (eficiencia). — [paper BIRD](https://arxiv.org/abs/2305.03111v2); [Technology Magazine](https://technologymagazine.com/news/huawei-datagallery-tops-benchmark-for-enterprise-text2sql-ai)
- El paper de capa semántica usa protocolo pareado (mismas 100 preguntas con/sin contexto semántico) y pruebas de significancia. — [arXiv 2604.25149](https://arxiv.org/abs/2604.25149)
- Ramp reporta accuracy evaluada por LLM-as-judge (99%), métrica más débil que etiquetas humanas. — [ZenML](https://zenml.io/llmops-database/ai-powered-merchant-classification-correction-agent)

### Inferencias
- **Eval de categorización:** 300–1,000 ítems de facturas reales de 5–10 pymes de rubros distintos (restaurante, bodega, taller), etiquetados por el dueño/contador; split temporal (entrenar memoria con meses anteriores, evaluar el último mes); reportar por separado: proveedor conocido, proveedor nuevo, tenant nuevo (cold start). Métricas: Top-1, Top-3, balanced accuracy, accuracy a nivel de *monto* (errores en ítems grandes pesan más), cobertura vs precisión según umbral de confianza, costo por factura. En producción: tasa de corrección del usuario por semana como métrica norte.
- **Eval de consultas NL:** 100–200 preguntas reales en español peruano (incluye coloquialismos, "el mes pasado", "lo de limpieza", comparaciones, top-N proveedores, IGV), cada una con resultado esperado calculado por SQL de referencia sobre datos fijos de prueba; métricas: exactitud del número (tolerancia 0), exactitud de la interpretación (categorías/periodo elegidos), tasa de abstención correcta en preguntas fuera de alcance, y un set adversarial (intento de leer otro tenant, inyección en descripciones de facturas) cuya métrica debe ser 0 fugas.
- Correr el eval como test de regresión en CI ante cada cambio de prompt/modelo; las cifras del paper semántico sugieren que mejorar el contexto/capa semántica rinde más que cambiar de modelo.

### Gaps
- No encontré guías publicadas específicas de evals para asistentes financieros de pymes; lo anterior es diseño inferido de los benchmarks citados.

## 6. Costos por factura con modelos actuales (precios oficiales, con fecha)

### Takeaway
Categorizar/extraer una factura con un modelo económico cuesta del orden de US$0.0002–0.003 por factura; para una pyme con 1,000 facturas/mes eso es centavos a ~3 dólares al mes. El costo de LLM no es la restricción a esta escala; la precisión y la UX de corrección sí.

### Cited Findings (precios por millón de tokens, entrada / salida, tier estándar)
- **Google Gemini API** (página "Last updated 2026-10-22"): Gemini 2.5 Flash-Lite $0.10 / $0.40; Gemini 3.1 Flash-Lite $0.25 / $1.50; Gemini 3.5 Flash-Lite $0.30 / $2.50; Gemini 2.5 Flash $0.30 / $2.50; Gemini 3 Flash Preview $0.50 / $3.00; Gemini 3.6/3.8 Flash $0.75 / $3.75 hasta el 31-dic-2026 (sube a $1.50 / $7.50 desde 1-ene-2027). Batch: 50% de descuento. — [Google AI pricing](https://ai.google.dev/gemini-api/docs/pricing) (consultado 2026-10-25)
- **Anthropic:** Claude Haiku 5.5 $0.10 / $0.50 (prompts ≤100K tokens; $0.50 / $2.50 por encima), cache hit $0.01; Claude Haiku 4.5 $1 / $5; Claude Sonnet 4.6 $3 / $15. Batch API 50% de descuento (Haiku 5.5 batch $0.05 / $0.25). Multiplicador 1.1x si se fuerza inferencia solo-EE.UU. — [Claude pricing](https://platform.claude.com/docs/en/about-claude/pricing) (consultado 2026-10-25)
- **OpenAI:** gpt-5-nano $0.05 / $0.40; gpt-5-mini $0.25 / $2.00; gpt-5.4-nano $0.20 / $1.25; gpt-5.4-mini $0.75 / $4.50; gpt-4.1-nano $0.10 / $0.40; gpt-4o-mini $0.15 / $0.60. Batch 50% de descuento. — [OpenAI pricing](https://developers.openai.com/api/docs/pricing) (consultado 2026-10-25; openai.com/api/pricing devolvió 403)
- Referencia empírica: categorizar un ítem con Claude Sonnet costó ~$0.004–0.011 por llamada (421–1,212 tokens de entrada, ~25 de salida). — [arXiv 2604.01615](https://arxiv.org/html/2604.01615v1)

### Inferencias (cálculos propios, no medidos)
- Supuesto: categorizar todas las líneas de una factura UBL ya parseada (XML → JSON compacto) en una llamada ≈ 1,500 tokens de entrada (prompt + lista de categorías del tenant + ítems + 3–5 ejemplos recuperados) y ≈ 200 de salida:
  - Gemini 2.5 Flash-Lite: ~$0.00023/factura; gpt-5-nano: ~$0.00016; Claude Haiku 5.5: ~$0.00025; gpt-5-mini: ~$0.00078; Gemini 3.1 Flash-Lite: ~$0.00068; Claude Haiku 4.5: ~$0.0025; Sonnet 4.6: ~$0.0075.
  - Con prompt caching del bloque fijo (instrucciones + categorías) y/o Batch API (categorización no necesita ser en tiempo real), el costo baja ~50% o más.
- Facturas peruanas electrónicas llegan como XML UBL estructurado: la *extracción* no requiere LLM (parseo determinista); el LLM solo hace falta para categorizar o para comprobantes en papel/foto/PDF (OCR multimodal: más tokens de entrada por imagen, posiblemente 2–5x el costo anterior; no lo medí).
- Una consulta NL vía tool-calling (≈2,000–4,000 tokens con definiciones de herramientas) cuesta del orden de $0.0005–0.003 con modelos económicos; despreciable frente al valor.
- Con la capa de memoria/kNN, la fracción de ítems que necesita LLM debería caer con el tiempo (el caso Mercado Libre muestra la ruta de destilar a modelos pequeños solo a escalas de cientos de millones, irrelevante para pymes).

### Gaps
- No medí tokens reales de un XML UBL de SUNAT típico ni de imágenes de recibos; las cifras por factura son estimaciones.
- Algunos modelos listados (p. ej. Gemini 3.x, Haiku 5.5) son recientes; no encontré evaluaciones independientes de su precisión en categorización en español.
