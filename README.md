# ============================================================
#          Created with: SDATA Microservice Framework
# ============================================================

Nombre de Compañía/Empresa: sdata
Nombre del Microservicio:   ms-plan
Nombre de la Base de Datos: plans
Nombre del Modelo/Protobuf: Plan
Usuario de GITHUB:          SDATA-Colombia
Versión de Protobuf:        v0.1.24

==============================================================
#  Despliegue en K3s
==============================================================

1. Construir imagen Docker:
   make build					
2. Importar imagen a K3s:
   make load
3. Aplicar manifiestos de Kubernetes:
   make apply
4. Verificar estado del Deployment:
   make status
5. Verificar logs del microservicio:
   make logs

Actualizar kong:

1. docker build -t docker.io/library/kong-grpc:3.9.3 -f infra/kong/Dockerfile .
2. docker save docker.io/library/kong-grpc:3.9.3 | sudo k3s ctr images import -
3. kubectl rollout restart deployment/kong -n sdata
4. kubectl rollout status deployment/kong -n sdata
5. kubectl exec -it deploy/kong -n sdata -- grep "host:" /etc/kong/kong.yaml

oscar@Debian:~/Documentos/mss-sdata/services/ms-plan$ kubectl describe nodes

==============================================================
#  Tipología(s) de Arquitectura Aplicada
==============================================================

La arquitectura que se ha aplicado se llama Arquitectura de Microservicios Basada en Eventos y Dirigida por Dominios (Event-Driven Microservices Architecture), combinada internamente con un patrón de Arquitectura por Capas Limpias (Clean Layered Architecture).
Aquí se detalla los tres nombres técnicos exactos que definen perfectamente el diseño:

## 1. Arquitectura de Microservicios Basada en Eventos (Event-Driven Architecture - EDA)
Se define así porque la comunicación no es solo directa (sincrónica). Al incluir Apache Pulsar (pkg/pulsar), los servicios se pueden comunicar emitiendo y reaccionando a eventos. Esto permite que el sistema sea reactivo, asíncrono y altamente desacoplado.

## 2. Arquitectura Orientada a Contratos (Design by Contract / API-First)
Al centralizar los Protobuf en un repositorio independiente de GitHub que todos importan, se aplicó un enfoque de diseño donde el contrato (la interfaz gRPC) manda. Ningún microservicio puede romper la comunicación porque todos respetan el mismo contrato versionado externamente.

## 3. Arquitectura de Capas Limpias (Clean Layered) con CQRS Híbrido
Dentro del código (internal/), se aplicó una división de responsabilidades muy estricta:

* Controller (Capa de Transporte): Recibe el gRPC (pkg/grpc).
* Mapper: Transforma los datos del contrato exterior a tus entidades internas.
* Service (Capa de Negocio): Modifica la lógica usando caché rápida (Dragonfly) y persistencia sólida (MongoDB).
* Repository (Capa de Datos): Abstrae las consultas a la base de datos.

Es un diseño de nivel empresarial (Enterprise-grade), ideal para sistemas que necesitan procesar millones de transacciones con latencia mínima.

==============================================================
#  Tests Unitarios y de Integración (grpcurl - CRUD)
==============================================================

# Desde la terminal
kubectl port-forward svc/ms-plan 50051:50051 -n sdata

# En otra terminal
# ------------------------------------------------------------

# Consultar la estructura del servicio gRPC (Crear):
grpcurl -plaintext localhost:50051 describe plan.v1.CreatePlanReq

# Crear un registro:
grpcurl -plaintext -d '{
  "data": {
    "name": "Óscar Darío Rendón Zuluaga",
    "description": "Primer registro desde grpcurl para probar el microservicio ms-plan",
    "status": true,
    "age": 30,
    "total": 5000,
    "price": 199.99,
    "tags": ["golang", "grpc", "mongodb", "pulsar", "dragonfly"],
    "numbers": [1,2,3,4,5],
    "metadata": {
      "country": "Colombia",
      "city": "Armenia"
    },
    "created_by": "oscar"
  }
}' localhost:50051 plan.v1.PlanService/CreatePlan

# ------------------------------------------------------------

# Consultar la estructura del servicio gRPC (Leer registro):
grpcurl -plaintext localhost:50051 describe plan.v1.GetPlanReq

# Consultar un registro:
grpcurl -plaintext -d '{
  "id":"6a63902601b32621d5c981e2"
}' localhost:50051 plan.v1.PlanService/GetPlan

# Consultar la estructura del servicio gRPC (Listar registros):
grpcurl -plaintext localhost:50051 describe plan.v1.ListPlanReq

# Consultar varios registros (paginación):
grpcurl -plaintext -d '{
  "page":1,
  "limit":10,
  "search":""
}' localhost:50051 plan.v1.PlanService/ListPlan

# ------------------------------------------------------------

# Consultar la estructura del servicio gRPC (Actualizar):
grpcurl -plaintext localhost:50051 describe plan.v1.UpdatePlanReq

# Actualizar un registro:
grpcurl -plaintext -d '{
  "id":"6a63902601b32621d5c981e2",
  "data": {
    "name": "Óscar Darío Rendón Zuluaga",
    "description": "Primer registro desde grpcurl para probar el microservicio ms-plan",
    "status": false,
    "age": 31,
    "total": 6000,
    "price": 299.99,
    "tags": ["golang", "grpc", "mongodb", "pulsar", "dragonfly"],
    "numbers": [1,2,3,4,5],
    "metadata": {
      "country": "Colombia",
      "city": "Armenia"
    },
    "created_by": "oscar"
  }
}' localhost:50051 plan.v1.PlanService/UpdatePlan

# ------------------------------------------------------------

# Consultar la estructura del servicio gRPC (Eliminar):
grpcurl -plaintext localhost:50051 describe plan.v1.DeletePlanReq

# Eliminar un registro:
grpcurl -plaintext -d '{
  "id":"6a63902601b32621d5c981e2"
}' localhost:50051 plan.v1.PlanService/DeletePlan

==============================================================
#  Tests Unitarios y de Integración (curl - CRUD)
==============================================================

# Desde la terminal
# ------------------------------------------------------------

# CreatePlan
curl -i -X POST http://192.168.0.240/api/v1/plan -H "Content-Type: application/json" -d '{
  "data": {
    "name": "Oscar",
    "description": "Primer registro desde http",
    "status": true,
    "age": 30,
    "total": 5000,
    "price": 199.99,
    "tags": [
      "golang",
      "grpc",
      "mongodb"
    ],
    "numbers": [
      1,
      2,
      3
    ],
    "metadata": {
      "country": "Colombia",
      "city": "Armenia"
    },
    "created_by": "oscar"
  }
}'

# ------------------------------------------------------------

# GetPlan
curl -i http://192.168.0.240/api/v1/plan/6a6a2e4949a9600e19d281b8

# ListPlan
curl -i http://192.168.0.240/api/v1/plans?page=1\&limit=10\&search=

# ------------------------------------------------------------

# UpdatePlan
curl -i -X PUT http://192.168.0.240/api/v1/plan/6a6b885d57a12e5176ba7e82 -H "Content-Type: application/json" -d '{
  "id":"6a6b885d57a12e5176ba7e82",
  "data":{
    "name":"Oscar Gómez",
    "description":"Registro actualizado",
    "status":true,
    "age":31,
    "total":9000,
    "price":250.75,
    "tags":[
      "grpc",
      "go",
      "kubernetes"
    ],
    "numbers":[
      10,
      20,
      30
    ],
    "metadata":{
      "updated":"true"
    },
    "created_by":"oscar"
  }
}'

# ------------------------------------------------------------

# DeletePlan
curl -i -X DELETE http://192.168.0.240/api/v1/plan/6a6a2e4949a9600e19d281b8
