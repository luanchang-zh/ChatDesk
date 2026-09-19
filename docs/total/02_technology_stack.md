# 个人 AI Workspace｜技术选型与依赖

> 文档版本：1.3  
> 文档定位：只说明项目采用什么技术、框架和基础依赖，以及各自负责什么。  
> 核心目标：非基础设施进程只有 Go 和 Python，不是微服务；四路检索（Direct / RAG / Graph / Web）由 Go 调度；Sandbox 属于 P2，不在当前栈。  
> 1.2 变更：取消 M0～M4 分期；图谱与 Neo4j 纳入当前技术栈，仍在同一 Python 进程内。  
> 1.3 变更：把 Go / Python 写成两个进程，把 PostgreSQL / Neo4j 等写成基础设施，避免读成微服务。

---

## 1. 总体技术路线

对前端和用户，系统只有一个 HTTP 入口。非基础设施进程固定为两个——Go 和 Python——这不是微服务。

PostgreSQL、Neo4j、本地文件目录、LLM Provider、Tavily 是基础设施或外部依赖，不计入这两个进程。

```text
Web Frontend
    ↓
Go 进程             ← 唯一对浏览器暴露
    ├── Eino：LLM 接入与 Workflow 编排
    ├── Direct Search：对标准化文本做精确搜索
    ├── Web Search：联网检索
    └── 本机调用 Python 进程
            ↓
      Python 进程
            ├── Parse / Normalize
            ├── RAG（LlamaIndex + pgvector）
            └── Graph（同一进程内的模块，不另起进程）
```

Go 掌握会话、上下文、检索调度、证据融合、工具权限和最终回答生成。Python 只做知识加工与本地知识检索，不对浏览器暴露，也不负责 Chat。

普通 RAG **不采用 RAGFlow**。本项目不需要它自带的 Chat、Agent、复杂工作台和完整知识库平台。

不把解析、RAG、图谱拆成三个进程。它们操作的是同一份标准化文本，拆开只会让文件和版本对不齐。

Neo4j 是图存储，不是第三个产品进程。Sandbox 属于 P2，不是当前技术栈的一部分。

---

## 2. 前端

### React + TypeScript

用于实现类似 ChatGPT 的网页工作台，包括：

- Chat 页面
- 会话列表
- Project / Knowledge 页面
- 文件上传
- 引用来源展示
- Tool Call 状态展示
- Graph 可视化
- 流式回答

采用：

```text
React
TypeScript
Vite
Tailwind CSS
shadcn/ui
TanStack Query
react-markdown
```

图谱展示使用：

```text
Cytoscape.js
```

前端不直接访问模型、Python 进程、Neo4j、pgvector 或 Web Search Provider。所有能力统一经过 Go 进程。

---

## 3. Go 进程

采用：

```text
Go
Gin
PostgreSQL
pgx
sqlc
goose
Eino
OpenTelemetry
```

HTTP 框架选定 **Gin**：资料多，个人项目开发更快。不再并行维护 Chi。

Go 进程负责：

```text
用户与权限
Project
Conversation
Message
会话分支
Run 状态
Context Builder
Token Budget
Retrieval Planner
Evidence Fusion
Citation 校验
Tool Registry
Workflow / P1 Agent Runtime
SSE
Trace
调用 Python 进程
```

普通问答使用固定 Retrieval Workflow。Agent Loop 属于 P1，不作为完整首版的前置依赖。

---

## 4. Eino

Eino 作为项目的 **LLM 接入与 AI 编排层**，运行在 Go 进程内。

负责：

```text
ChatModel
Streaming
Tool Calling
Workflow
Graph / Chain
Callback / Trace
P1 Agent Loop
```

不负责：

```text
Conversation 持久化
用户权限
文档生命周期
业务数据库
知识库版本管理
文件解析
向量存储
```

项目自身定义 Model、Tool、Retriever 等业务接口，在内部通过 Eino 适配实际模型。Go 的 Retriever 接口既可走本地 Direct / Web，也可调用 Python 进程；上层 Planner 不感知 LlamaIndex 或 Neo4j。

建议支持至少：

```text
OpenAI-compatible Chat Model
OpenAI-compatible Embedding
DeepSeek / Qwen 等兼容 Provider
```

模型选择对上层 Conversation 和 Retrieval 模块透明。

---

## 5. Python 进程

第二个非基础设施进程。内部是知识作业流水线：解析、RAG、图谱都在这一个进程里。

FastAPI 只是 Go 在本机调用它的方式，不是对外产品 API，也不是按能力拆开的微服务。

采用：

```text
Python
FastAPI
Docling
LlamaIndex
PostgreSQL + pgvector
neo4j-graphrag
Neo4j
```

当前要跑起来的东西分成两类：

```text
非基础设施进程：Go、Python
基础设施：PostgreSQL、Neo4j、本地文件目录
```

图谱与语义检索共用这一个 Python 进程。

### 5.1 为什么保留 Python

本项目真正想练的是：

```text
Go 进程
Retrieval Planner
Context Engineering
会话系统
引用核验
```

而不是从零实现完整的：

```text
复杂文档解析
Chunk Pipeline
Vector Retriever
Rerank Adapter
实体关系抽取流水线
```

LlamaIndex 和 Docling 只作为 Python 进程内部的薄层。Chat、权限、最终回答不交给它们。

### 5.2 Python 进程负责

```text
读取原始文件
解析为标准化文本
写入规范化产物
Chunk
Embedding
Vector Retrieval
可选 BM25 / Hybrid / Rerank
实体关系抽取与图查询
```

对 Go 只提供本机调用入口，例如：

```text
POST /index
POST /retrieve
DELETE /documents/:id
POST /graph/build
POST /graph/query
```

Go 传入文档 ID、用户范围和版本，不让 Python 自己决定产品权限。最终回答仍然由 Go + Eino 完成。

### 5.3 Vector Store

默认：

```text
同一套 PostgreSQL + pgvector
```

建议分 schema：Go 业务数据与知识向量分开，避免混表。个人项目不额外部署独立 Vector Database。

如果以后数据量明显增大，可以替换为 Qdrant；Go 上层仍然只调用 Python 进程，不感知具体 Vector Store。

---

## 6. 文档解析

解析是 Python 进程的第一段作业，不是第三个进程。

P0 支持：

```text
Markdown
UTF-8 TXT
可提取文本的 PDF
```

P1 再支持：

```text
DOCX
扫描 PDF 与更复杂版式
```

解析器默认使用 **Docling**。不并行引入 MinerU，除非 Docling 无法覆盖后续某类文件。

输出必须是一份带定位元数据的标准化 Markdown / Text，供 Direct、RAG、Graph 共用。PDF 页码或阅读顺序无法保留时，元数据要如实记录，禁止编造原文件行号。

---

## 7. Direct Search

Direct Search 用于解决：

```text
精确字符串
函数名
错误码
文件名
受限正则
代码符号
```

语料是 Python 进程写出的**标准化文本**，不是原始 PDF 二进制，也不是宿主机任意目录。

实现上由 Go 通过受限参数调用：

```text
ripgrep
```

不允许模型自由拼接 Shell 命令。内部抽象：

```text
DirectRetriever
    ↓
标准化文本（FileStore）
    ↓
ripgrep
    ↓
Document / Chunk Location
```

后续需要 BM25 时，优先用 PostgreSQL Full Text Search，避免再部署 Elasticsearch。

---

## 8. 知识图谱

图谱是 Python 进程内部的模块，沿用同一进程和同一份标准化文本。不另起进程。

采用：

```text
neo4j-graphrag
Neo4j
```

负责：

```text
Entity Extraction
Relation Extraction
Entity Resolution
Knowledge Graph Build
Entity Search
Graph Traversal
Multi-hop Retrieval
Graph → Source Chunk 回溯
```

Go 只通过统一 GraphRetriever 调用 Python 进程。图谱失败不得阻断 Direct / RAG。

Neo4j 主要保存：

```text
Entity
Relation
Source / Provenance
Graph Path
```

不再在 Neo4j 里重复维护一套完整的 Chunk Vector RAG。如果需要实体链接，可以只为 Entity 建立 embedding。

---

## 9. Web Search

Web Search 作为 Go 内的独立 Tool / Retriever，不经过 Python。

先封装：

```text
WebSearchProvider
```

第一版默认：

```text
Tavily
```

Brave / Exa / SearXNG 只作为可替换实现，不在第一版同时接入。

能力分成：

```text
web_search
web_fetch
```

### web_search

负责返回：

```text
Title
URL
Snippet
PublishedAt
Source
```

### web_fetch

由 Go 访问指定公开网页并抽取正文，需要：

```text
URL 校验
超时
响应大小限制
内网地址限制
重定向限制
正文清洗
```

防止 SSRF。产品默认关闭联网，由用户按会话授权。

---

## 10. PostgreSQL

第一版只需要一套 PostgreSQL。

Go 保存：

```text
User
Project
Conversation
Message
Run
Tool Call
Knowledge Metadata
Document Metadata
Citation
Trace Metadata
```

Python 进程保存：

```text
Chunk
Embedding（pgvector）
检索用辅助表
```

数据库访问：

```text
Go：pgx + sqlc + goose
Python：仅访问知识 schema，不写会话表
```

---

## 11. 文件存储

个人开发阶段使用本地目录，Go 与 Python 进程挂载同一数据卷：

```text
/data/documents     原始文件，Go 写入
/data/normalized    标准化文本，Python 进程写入
/data/artifacts     可选产物
```

代码层定义：

```text
FileStore Interface
```

后续可替换 MinIO / S3 / OSS。业务代码不直接依赖某个对象存储 SDK。

契约：

```text
Go 写入原始文件并登记文档版本
Go 请求 Python 进程 index
Python 进程读取原始文件，写出标准化文本，再建 RAG / 图谱索引
Direct Search 只扫描标准化文本
```

---

## 12. 会话流式输出

Chat 使用：

```text
HTTP + SSE
```

而不是默认使用 WebSocket。

主要流式事件：

```text
message.delta
tool.started
tool.completed
retrieval.started
retrieval.completed
citation
run.completed
run.failed
```

WebSocket 仅在以后真的出现双向高频实时通信需求时再引入。

---

## 13. 可观测性

首版：

```text
slog
OpenTelemetry
```

主要记录：

```text
Request Trace
LLM latency
TTFT
Token Usage
Retrieval latency
Retriever route
Retrieved Evidence
Python 作业状态
Tool latency
Errors
```

以后可以接 Jaeger / Tempo / Prometheus / Grafana / Langfuse，但不作为 MVP 必须依赖。

---

## 14. Sandbox（P2，不在当前栈）

主体完成前不引入第三个执行进程。本文保留约束，避免以后范围漂移。

若启用，必须与 Go / Python 这两个进程分开，只提供一种预置语言（Python 或 Go），不提供不受限 Shell。

基本限制：

```text
CPU / Memory / PID / Timeout / Disk Limit
Network Disabled by Default
Non-root
No Docker Socket
No Host FS Mount
```

---

## 15. 当前技术栈

```text
Frontend
├── React / TypeScript / Vite
├── Tailwind CSS / shadcn/ui / TanStack Query
├── react-markdown
└── Cytoscape.js

Go 进程
├── Go
├── Gin
├── Eino
├── pgx / sqlc / goose
├── ripgrep（Direct Search）
└── Tavily（Web Search，默认关闭）

Python 进程
├── FastAPI（仅供 Go 本机调用）
├── Docling
├── LlamaIndex
└── neo4j-graphrag

基础设施
├── PostgreSQL + pgvector
├── Neo4j
└── 本地 FileStore
```

Agent Loop 属于 P1，不进入当前默认依赖。Sandbox 属于 P2，不进入当前栈。

---

## 16. 技术边界总结

```text
Go 进程
→ 唯一对浏览器暴露：会话、权限、Planner、Context、Citation、SSE

Eino
→ Go 进程内的 LLM 与 Workflow

Python 进程
→ 解析、语义索引、图谱；不对浏览器暴露

LlamaIndex / Docling
→ Python 进程内部库，不接管 Chat

Neo4j
→ 图存储，是基础设施，不是产品进程

ripgrep
→ 只搜标准化文本

Tavily
→ 只负责互联网搜索结果；正文读取由 Go 完成
```

因此项目不会变成某个 AI 框架的套壳，也不会膨胀成微服务。非基础设施进程始终只有 Go 和 Python。

真正由项目自身实现的核心仍然是：

```text
Conversation
Context Engine
Retrieval Planner
Evidence Fusion
Citation
权限
Trace
```
