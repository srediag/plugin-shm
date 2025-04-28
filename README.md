# Plugin-SHM

**plugin-shm** is a high-performance shared memory plugin runtime, forked from [`shmipc-go`](https://github.com/srediag/plugin-shm), and redesigned for secure, observable, and extensible plugin execution.

Built for **SREDIAG** and similar projects, it provides ultra-low-latency communication between plugins and hosts, governance, hot-reload, crash isolation, and optional TCP transport for distributed plugin support.

---

## Key Features

- **Zero-copy shared memory communication** between host and plugins
- **Hot-reload support** without service downtime
- **Secure plugin loading** (SHA256 + cosign signature validation)
- **Heartbeat and liveness monitoring** of plugins
- **Governance audit logs** for plugin operations
- **Planned multi-platform support** (Linux, MacOS, Windows)
- **Future distributed plugin mode** (via TCP/QUIC)

---

## Quick Start

```bash
git clone https://github.com/srediag/plugin-shm.git
cd plugin-shm
make build
