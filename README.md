# nanolambda

> the intelligent serverless platform
> aws lambda meets facebook prophet

nanolambda is a research-grade serverless platform that uses ai to predict traffic spikes and pre-warm containers, eliminating cold starts.

## features

- **scale-to-zero:** containers automatically stop when idle to save resources.
- **ai prediction:** uses time-series forecasting (prophet) to predict when users will arrive.
- **pre-warming:** starts containers *before* traffic hits, ensuring instant responses.
- **developer experience:** polished cli (`nanolambda`) for deploying and managing functions.

## architecture

- **cli (go):** developer tools for deploying and monitoring.
- **gateway (go):** reverse proxy, container manager, and load balancer.
- **intelligence (python):** prophet model + prometheus metrics.
- **runtime (docker):** isolated execution environments for python functions.

## quick start

### prerequisites
- docker & docker compose
- go 1.21+
- python 3.10+

### 1. start the platform
```bash
# start support services (prometheus, ai)
docker-compose up -d

# build the base runtime image (one time)
docker build -t nanolambda/base-python:3.11 runtime/python

# start the gateway (keep this running)
# in a new terminal:
go run ./cmd/gateway
# or if compiled:
# .\gateway.exe
```

### 2. deploy a function
```bash
# build the cli
go build -o nanolambda.exe ./cmd/cli

# initialize a new function
.\nanolambda.exe init hello-world

# deploy it
.\nanolambda.exe deploy hello-world
```

### 3. invoke it
```bash
# cold start (first time ~2s)
curl -x post http://localhost:8080/function/hello-world -d '{"name": "developer"}'

# hot start (second time ~50ms)
curl -x post http://localhost:8080/function/hello-world -d '{"name": "developer"}'
```

### 4. watch the magic
open the dashboard to see real-time metrics:
```bash
# start dashboard server (in new terminal)
python web/dashboard/serve.py

# open in browser
.\nanolambda.exe dashboard
```

## how the ai works
1. **collect:** prometheus scrapes traffic metrics every 5s.
2. **learn:** prophet trains on the last 7 days of history (simulated in demo).
3. **predict:** every 5 mins, it forecasts the next 10 mins.
4. **act:** if a spike is predicted, it calls the gateway to pre-warm containers.

## project structure
```
├── cmd/
│   ├── cli/        # nanolambda command
│   └── gateway/    # the main server
├── pkg/            # shared go libraries
├── prophet/        # python ai service
├── runtime/        # function runners
└── web/            # dashboard
```