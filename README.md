# 🚀 URL Shortener


[![Go](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![Redis](https://img.shields.io/badge/Redis-Cache-DD0031?logo=redis&logoColor=white)](https://redis.io)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-Database-4169E1?logo=postgresql&logoColor=white)](https://www.postgresql.org)
[![Kafka](https://img.shields.io/badge/Kafka-Event_Stream-231F20?logo=apache-kafka&logoColor=white)](https://kafka.apache.org)

A lightweight but scalable **Go** pet project exploring backend performance and scalability.  
It uses **Redis** for fast caching, **PostgreSQL** as the main datastore, and **Kafka** for processing click events.  
Designed to handle high throughput, it’s been tested with **bombardier** (up to millions of requests) to explore concurrency, rate limiting, and real-world latency under load.  
Metrics are exposed through **Prometheus** and visualized with **Grafana**.

---

### 🧩 Architecture Overview

<p align="center">
  <img src="./shortener.png" alt="System Architecture" width="800"><br>
  <em>System flow: middleware, caching, event streaming, and observability</em>
</p>

---

### ⚙️ Key Components
- 🧠 **Middleware:** Handles logging, rate limiting, and validation.  
- 🔗 **Redirect Handler:** Validates slugs, checks cache, and redirects requests.  
- ⚡ **Redis:** Caches active URLs and metrics for low latency.  
- 🗃️ **PostgreSQL:** Persistent storage and source of truth.  
- 📬 **Kafka:** Manages click event streaming and background flushing.  
- 📊 **Prometheus & Grafana:** Observability and performance metrics.  
- ⏱️ **Scheduler:** Periodic jobs for cleanup and metric aggregation.

---

### 🧠 Highlights
- 🧨 Tested with **tens of thousands to 1M+ requests** using **bombardier**  
- 🧵 Explores **concurrency**, **rate limiting**, and **load handling** in Go  
- 🧩 Built as a **learning project** to understand scalable backend design  

---
