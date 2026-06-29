# microservice-golang-architecture

## 🎯 Objetivo do Projeto

Desenvolver uma prova de conceito (PoC) de **arquitetura de microsserviços escalável**, utilizando Go para orquestrar a comunicação eficiente entre serviços via REST e gRPC.

O sistema foi arquitetado como um laboratório prático para testes de estresse computacional: um Gateway recebe requisições externas e as repassa internamente para um serviço de processamento criptográfico intensivo. O foco principal é demonstrar a conteinerização otimizada (Docker Multi-stage) e a validação do **Escalonamento Horizontal Automático (HPA) no Kubernetes** sob alta demanda de CPU.
