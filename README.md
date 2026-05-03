# Simplified Stock Market API

A high-availability, stateless REST API simulating a simplified stock exchange. Built with **Go**, **Redis**, and **Nginx**, and containerized using **Docker**.

### Getting Started

The application is fully containerized and requires only Docker to run. It supports Windows, Linux, and macOS.

### Prerequisites
- Docker & Docker Compose installed and running.

### How to run

The application can be started using a single command, passing the desired port as an argument.

**Windows:**
```cmd
.\start.bat 8080
```

**Linux / macOS:**
```bash
# Make sure the script is executable before the first run
chmod +x start.sh
./start.sh 8080
```

The API will be available at `http://localhost:8080` (or whichever port you specified).

---

### Architecture & Design Decisions

To meet the strict requirements of High Availability (HA), concurrency handling, and environmental agnostic setup, several architectural decisions were made:

### 1. High Availability (HA) & Load Balancing
The environment consists of 4 containers:
- **Nginx (Load Balancer)**: Acts as the entry point, distributing traffic across backend instances using a Round-Robin algorithm.
- **Go API Instances (api1 & api2)**: Two identical, stateless instances of the Go application.
- **Redis**: The centralized, persistent storage.

If the `/chaos` endpoint is triggered, it forcefully kills one of the API instances. Nginx immediately detects the connection drop and routes all subsequent traffic to the surviving instance, resulting in zero downtime (beyond a potential 502 error for the exact request that hit the killed instance). Docker automatically restarts the dead container in the background.

### 2. Why 2 Static Containers instead of Kubernetes?
While a production environment would utilize Kubernetes (K8s) or Docker Swarm with dynamic Pod scaling based on traffic, this project uses two statically defined containers (`api1`, `api2`) in `docker-compose.yml`. 
*Reasoning:* The assignment explicitly stated that no assumptions about the environment should be made apart from Docker being available. Relying on K8s/Minikube would violate this constraint. This static multi-container approach guarantees a 100% "plug-and-play" experience for the reviewer on any machine.

### 3. Concurrency & Optimistic Locking
To prevent race conditions during heavy concurrent trading, the application uses **Optimistic Locking** via Redis `WATCH` and `MULTI/EXEC` transactions. 
If a stock's quantity in the bank changes between the time a user checks the availability and attempts to buy, the transaction fails and the API handles it gracefully without causing a negative balance.

### 4. Redis Pipeline vs. TxPipelined
Two different pipelining strategies were used based on the context:
- **`Pipeline()` in `POST /stocks`:** Batches commands purely for network optimization, reducing latency without blocking the Redis thread during initialization.
- **`TxPipelined()` in `POST /wallets/...`:** Wraps commands in a Redis Transaction (`MULTI/EXEC`) to provide strict ACID guarantees, ensuring trades are fully atomic.

---

## 📡 API Endpoints

| Method | Endpoint | Description |
| :--- | :--- | :--- |
| `POST` | `/stocks` | Sets the initial state of the bank. (Accepts: `{"stocks": [{"name": "AAPL", "quantity": 100}]}`) |
| `GET` | `/stocks` | Returns the current stock inventory of the bank. |
| `POST` | `/wallets/{id}/stocks/{name}` | Simulates buying/selling a stock. (Accepts: `{"type": "buy"}` or `{"type": "sell"}`) |
| `GET` | `/wallets/{id}` | Returns the current state of a specific wallet. |
| `GET` | `/wallets/{id}/stocks/{name}` | Returns the exact quantity of a specific stock in a wallet (Returns a raw number). |
| `GET` | `/log` | Returns the audit log of successful trades (Max 10,000 entries). |
| `POST` | `/chaos` | Simulates a fatal crash by killing the instance serving the request. |
