# Kubernetes

Manifests para executar o gateway HTTP, o processor gRPC e o Metrics Server.

## Aplicar

```sh
kubectl apply -k k8s
```

## Imagens

Os manifests usam imagens locais:

- `api-gateway:latest`
- `compute-processor:latest`

Em um registry real, altere `image` nos deployments antes do `kubectl apply`.

## Escalonamento

- Gateway: HPA por CPU e memoria com limites baixos de CPU, adequado para carga leve de I/O HTTP.
- Processor: HPA por CPU em 75%, com `requests.cpu` e `limits.cpu` iguais para reduzir variacao de latencia em carga CPU-bound.

RPS nao e exposto pelo Metrics Server. Para escalar por RPS, adicione um adapter de metricas customizadas, como Prometheus Adapter ou KEDA, e troque o HPA do gateway para usar `Pods`, `Object` ou `External` metrics.
