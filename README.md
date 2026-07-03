# 🚀 Microservice Golang Architecture (Compute Processor)

## 🏗️ Arquitetura

O projeto é dividido em dois microsserviços principais:

1. **API Gateway (Porta 8080):** - Atua como Ingress/BFF recebendo chamadas HTTP/JSON.
   - Faz validação semântica e estrutural (Fail-Fast).
   - Comunica-se via gRPC com o backend aplicando *Timeouts* e *Graceful Degradation*.
2. **Compute Processor (Porta 50051):** - Serviço gRPC focado em tarefas CPU-bound.
   - Utiliza *Worker Pools* para limitar concorrência e evitar *Context Switching* excessivo.
   - Implementa *Graceful Shutdown* e cancelamento de contexto (interrompe o cálculo se o cliente desconectar).

## 🛠️ Pré-requisitos

Para rodar o projeto localmente, você precisará de:
* [Docker](https://docs.docker.com/get-docker/)
* Um cluster Kubernetes local ([Minikube](https://minikube.sigs.k8s.io/docs/start/), [Kind](https://kind.sigs.k8s.io/) ou Docker Desktop com K8s habilitado)
* `kubectl` instalado e configurado

## 🚀 Como Rodar (Kubernetes)

### 1. Build das Imagens Docker
Se você não estiver baixando as imagens de um Registry (como o GHCR), faça o build localmente:

```bash
# Build do API Gateway
docker build -t api-gateway:latest -f api-gateway/Dockerfile .

# Build do Compute Processor
docker build -t compute-processor:latest -f compute-process/Dockerfile .
```
### 2. Aplicar o Metrics Server (Opcional, para o HPA funcionar)

O escalonamento automático (Horizontal Pod Autoscaler) depende do Metrics Server:

```bash
kubectl apply -f k8s/metrics-server.yaml
```

### 3. Fazer o Deploy da Aplicação

Aplique todos os manifestos de infraestrutura (Deployments, Services e HPAs):

```bash
# Se os arquivos estiverem em uma pasta k8s/
kubectl apply -f k8s/
```

### 4. Expor a Porta Localmente

Para testar a API na sua máquina local, faça o Port-Forward do serviço do Gateway:

```bash
kubectl port-forward svc/api-gateway 8080:80
```

## 🧪 Como Testar

Com a aplicação rodando e o port-forward ativo, você pode disparar uma requisição para a rota `/process`.

**Exemplo usando cURL:**

```bash
curl -X POST http://localhost:8080/process \
  -H "Content-Type: application/json" \
  -d '{
    "request_id": "req-12345",
    "complexity": 50000,
    "payload": "SGVsbG8gV29ybGQ="
  }'
```
**Resposta Esperada:**

O Gateway vai aguardar o processamento do worker gRPC e retornará o hash final e o tempo que levou:

```json
{
  "request_id": "req-12345",
  "result_hash": "...",
  "processing_time_ms": 142
}
```
