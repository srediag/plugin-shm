# Plugin-SHM Roadmap (Detailed)

This roadmap organizes the entire evolution of **plugin-shm**,  
marking each feature’s **priority**, **status**, and **origin**.

| Legend | Meaning |
|:------:|:--------|
| 🟩 | Inherited from fork (already functional) |
| 🟨 | Partially inherited, needs extension |
| 🟥 | New feature to be implemented |

---

## Phase 0 — Baseline Fork

| Item | Priority | Status | Origin |
|:-----|:---------|:-------|:------|
| Local shared memory transport (mmap) | Critical (P0) | 🟩 Ready (from shmipc-go) | Fork |
| Ringbuffer IPC (lockless) | Critical (P0) | 🟩 Ready (forked) | Fork |
| Basic session management | Critical (P0) | 🟩 Ready (forked) | Fork |
| Event dispatcher (epoll on Linux) | Important (P1) | 🟩 Ready (forked) | Fork |
| Basic error handling (retry, EOF) | Important (P1) | 🟩 Ready (forked) | Fork |

---

## Phase 1 — Core Plugin Runtime

| Item | Priority | Status | Origin |
|:-----|:---------|:-------|:------|
| Handshake Protocol (Name, Version, Capabilities) | Critical (P0) | 🟥 To implement | New |
| Plugin Heartbeat mechanism | Critical (P0) | 🟥 To implement | New |
| Plugin Manager (load, unload, reload) | Critical (P0) | 🟥 To implement | New |
| Audit Log (plugin operations: load/unload/fail) | Important (P1) | 🟥 To implement | New |
| Max Write Buffer limit per plugin stream | Important (P1) | 🟥 To implement | New |

---

## Phase 2 — Security Foundations

| Item | Priority | Status | Origin |
|:-----|:---------|:-------|:------|
| SHA256 checksum validation on plugin binary | Critical (P0) | 🟥 To implement | New |
| Cosign signature verification (optional) | Important (P1) | 🟥 To implement | New |
| Seccomp profiles per plugin process | Important (P2) | 🟥 To implement | New |
| Read-only filesystem mount for plugins | Important (P2) | 🟥 To implement | New |
| Memory budget enforcement (soft limit + SIGSTOP) | Important (P2) | 🟥 To implement | New |

---

## Phase 3 — Observability & Governance

| Item | Priority | Status | Origin |
|:-----|:---------|:-------|:------|
| Per-plugin metrics (message counts, latencies) | Important (P2) | 🟥 To implement | New |
| Plugin Status Tracking (Running, Draining, Failed) | Important (P2) | 🟥 To implement | New |
| Prometheus metrics exposure (future) | Optional (P3) | 🟥 To implement | New |

---

## Phase 4 — Multi-Platform Support

| Item | Priority | Status | Origin |
|:-----|:---------|:-------|:------|
| Linux full support | Critical (P0) | 🟩 Ready (inherited) | Fork |
| MacOS support (shared memory via shm_open fallback) | Important (P2) | 🟥 To implement | New |
| Windows support (CreateFileMapping) | Optional (P3) | 🟥 To implement | New |

---

## Phase 5 — Transport Abstraction

| Item | Priority | Status | Origin |
|:-----|:---------|:-------|:------|
| Define `Transport` interface (Send/Receive/Close) | Important (P1) | 🟥 To implement | New |
| TCP Transport (localhost, optional distributed mode) | Optional (P3) | 🟥 To implement | New |
| QUIC Transport (future) | Optional (P4) | 🟥 To design | New |

---

## Phase 6 — Resilience & Hot-Reload

| Item | Priority | Status | Origin |
|:-----|:---------|:-------|:------|
| Plugin Hot Reload (Drain → Swap → Unload old) | Critical (P1) | 🟥 To implement | New |
| Crash detection and auto-restart (basic) | Important (P2) | 🟥 To implement | New |
| Safe fallback when plugin stalls (SIGKILL, recover) | Important (P2) | 🟥 To implement | New |

---

## Phase 7 — Advanced Extensions

| Item | Priority | Status | Origin |
|:-----|:---------|:-------|:------|
| Timed synchronization mode (sleep + poll offline) | Optional (P4) | 🟥 To implement | New |
| Dynamic plugin discovery (distributed mode) | Future (P5) | 🟥 To design | New |
| Secure remote plugin delivery (cosigned bundle) | Future (P5) | 🟥 To design | New |

---

## Milestone Mapping

| Milestone | Target Items |
|:----------|:-------------|
| **v0.1** | Phase 0 + Phase 1 basic |
| **v0.2** | Phase 2 (Security) + Phase 3 (Governance) |
| **v0.3** | Phase 4 (Multi-platform) + Phase 5 (Transport abstraction) |
| **v0.4** | Phase 6 (Hot Reload + Crash Recovery) |
| **v1.0** | Phase 7 (Distributed + Dynamic Extension) |

---

## Licensing

Plugin-SHM is licensed under Apache 2.0.  
This project is a fork of [`cloudwego/shmipc-go`](https://github.com/srediag/plugin-shm), with major architectural upgrades focused on plugin governance, hot-reload, and security.
