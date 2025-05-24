# Stateless Libvirt REST API Node Controller

A lightweight, stateless Golang service exposing libvirt hypervisors over a simple REST API.  
Great as a building block for custom cloud control-planes, auto-scaling clusters or edge-compute node agents.

---

## Features

- Stateless: no local database, all state lives in libvirt
- Full libvirt domain lifecycle: define, start, stop, reboot, undefine
- Snapshot management
- Pluggable auth (e.g. Bearer token)e

---

## Prerequisites

- Go 1.18+  
- Libvirt (client libraries and daemon)  

## Quick Start

```bash
git clone https://github.com/UltraSive/libvirt-hypervisor-controller.git
cd libvirt-hypervisor-controller
chmod +x packages.sh
./packages.sh
make build

# Set your libvirt URI (e.g. qemu:///system)
export LIBVIRT_URI=qemu:///system
# Optional: listening address (default :8080)
export LISTEN_ADDR=:9090

./main
```

---

## Configuration

| ENV Variable     | Required | Default        | Description                             |
|------------------|----------|----------------|-----------------------------------------|
| NODE_ID          | false    | NODE_1         | The node ID for webhook events          |
| LIBVIRT_URI      | false    | qemu:///system | libvirt connection URI (required)       |
| PORT             | false    | 8080           | HTTP bind address                       |
| AUTH_TOKEN       | false    | —              | Static bearer token for simple auth     |
| WEBHOOK_ENDPOINT | false    | —              | HTTP endpoint for events                |
| CACHE_DIR        | false    | —              | Cache for VM image templates            |
| CACHE_SECONDS    | false    | —              | How long should VM images be cached     |

---

## API Reference

[Swagger OpenAPI Docs](https://ultrasive.github.io/hypervisor-api-docs)

---

## Webhook Events
When `WEBHOOK_ENDPOINT` is set, the controller will emit HTTP POST requests to the specified endpoint on major domain lifecycle changes. Each event contains the `domain_id` and `event_type`.

### Event Payload Format

```json
{
  "object": "event",
  "node_id": "NODE_1",
  "domain_id": "vm-123",
  "type": "domain.defined",
  "data": {},
  "timestamp": "2024-05-23T18:25:43.511Z"
}
```

### Event Types

| Event Type                | Description                   |
|---------------------------|-------------------------------|
| `domain.defined`          | Domain was defined (created)  |
| `domain.started`          | Domain was started            |
| `domain.stopped`          | Domain was gracefully stopped |
| `domain.shutdown`         | Domain shutdown was initiated |
| `domain.rebooted`         | Domain was rebooted           |
| `domain.undefined`        | Domain was deleted/undefined  |
| `domain.snapshot_created` | A snapshot was created        |
| `domain.snapshot_deleted` | A snapshot was deleted        |

---

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE) for details.
