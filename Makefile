# ============================================================
# SDATA Microservice: ms-plan
# Mantiene los comandos de Docker/Kubernetes y agrega calidad.
# ============================================================

.DEFAULT_GOAL := build

SERVICE := ms-plan
IMAGE := $(SERVICE):latest
NAMESPACE := sdata
TAR := $(SERVICE).tar

KUBECTL := sudo kubectl
CTR := sudo k3s ctr

# Pruebas unitarias cubren paquetes internos e infraestructura.
PKGS := ./internal/... ./pkg/...

# Token opcional: se puede exportar SONAR_TOKEN en el entorno.
SONAR_TOKEN ?=

.PHONY: build push save load images docker-prune \
        apply delete restart redeploy logs previous-logs \
        status all describe events top pvc shell clean help \
        test cover cover-all test-defects test-integration \
        lint sonar k6-grpc k6-rest smoke test-up test-down

# ============================================================
# Docker (original)
# ============================================================

build:
	docker build -t $(IMAGE) .

push:
	docker push $(IMAGE)

save:
	docker save $(IMAGE) -o $(TAR)

load: save
	$(CTR) images import $(TAR)
	rm -f $(TAR)

images:
	$(CTR) images ls | grep $(SERVICE)

docker-prune:
	docker builder prune -f
	docker image prune -f

# ============================================================
# Kubernetes (original)
# ============================================================

apply:
	$(KUBECTL) apply -f k8s/

delete:
	$(KUBECTL) delete -f k8s/

restart:
	$(KUBECTL) rollout restart deployment/$(SERVICE) -n $(NAMESPACE)

redeploy: build load apply
	$(KUBECTL) rollout status deployment/$(SERVICE) -n $(NAMESPACE)

# ============================================================
# Logs (original)
# ============================================================

logs:
	$(KUBECTL) logs deployment/$(SERVICE) -n $(NAMESPACE) -f

previous-logs:
	$(KUBECTL) logs deployment/$(SERVICE) -n $(NAMESPACE) --previous

# ============================================================
# Estado (original)
# ============================================================

status:
	$(KUBECTL) get pods -n $(NAMESPACE)

all:
	$(KUBECTL) get all -n $(NAMESPACE)

describe:
	$(KUBECTL) describe deployment/$(SERVICE) -n $(NAMESPACE)

events:
	$(KUBECTL) get events -n $(NAMESPACE) --sort-by=.metadata.creationTimestamp

# ============================================================
# Recursos (original)
# ============================================================

top:
	$(KUBECTL) top pod -n $(NAMESPACE)

pvc:
	$(KUBECTL) get pvc -n $(NAMESPACE)

# ============================================================
# Shell (original)
# ============================================================

shell:
	$(KUBECTL) exec -it deployment/$(SERVICE) -n $(NAMESPACE) -- sh

# ============================================================
# Pruebas unitarias y cobertura
# ============================================================

# Solo pruebas normales, sin activar la etiqueta integration ni servicios externos.
test:
	go test -race -count=1 -v $(PKGS)

# Reporte de cobertura unitaria para Go y SonarQube.
cover:
	go test -count=1 -coverprofile=coverage.out -json $(PKGS) > test-report.json
	go tool cover -func=coverage.out | tail -n 1
	go tool cover -html=coverage.out -o coverage.html

# Cobertura combinada: requiere servicios de prueba configurados y disponibles.
# No usar contra una base de datos de desarrollo o producción.
cover-all:
	MSPLAN_IT=1 go test -count=1 -tags integration \
		-coverpkg=./internal/...,./pkg/... -coverprofile=coverage.out -json \
		./internal/... ./tests/... > test-report.json
	go tool cover -func=coverage.out | tail -n 1
	go tool cover -html=coverage.out -o coverage.html

# Ejecuta las pruebas etiquetadas como defectos conocidos.
# Están diseñadas para fallar mientras los defectos documentados existan.
# La suite original usa '|| true' para continuar aunque fallen; aquí se conserva
# ese comportamiento, pero se advierte que el código de salida NO es un veredicto.
test-defects:
	-go test -count=1 -v -tags defects ./internal/...
	-MSPLAN_IT=1 MSPLAN_IT_DEFECTS=1 go test -count=1 -v -tags integration ./tests/... -run Defect

# Pruebas reales de integración: confirmar primero que la configuración apunta
# exclusivamente a MongoDB/Dragonfly/Pulsar de pruebas.
test-integration:
	MSPLAN_IT=1 go test -count=1 -v -tags integration ./tests/...

# ============================================================
# Análisis estático
# ============================================================

lint:
	golangci-lint run ./...

sonar: cover
	sonar-scanner -Dsonar.token="$(SONAR_TOKEN)"

# ============================================================
# Rendimiento (requiere servicio accesible y port-forward/config adecuados)
# ============================================================

k6-grpc:
	k6 run --summary-export=k6-grpc-summary.json k6/plan_grpc_load.js

k6-rest:
	k6 run --summary-export=k6-rest-summary.json k6/plan_rest_kong.js

# Contrato REST contra Kong. Ejemplo: BASE_URL=https://192.168.0.240 make smoke
smoke:
	./scripts/kong_smoke.sh

# ============================================================
# Infraestructura LOCAL para pruebas (separada de Kubernetes)
# ============================================================

# Levanta el compose de pruebas; confirmar primero variables/puertos.
test-up:
	docker compose -f docker-compose.test.yml up -d

# Detiene contenedores de prueba SIN borrar volúmenes persistentes.
# No agrega -v deliberadamente para reducir el riesgo de borrar datos.
test-down:
	docker compose -f docker-compose.test.yml down

# ============================================================
# Limpieza (original)
# ============================================================

clean:
	rm -f $(TAR)

# ============================================================
# Ayuda
# ============================================================

help:
	@echo ""
	@echo "Docker/Kubernetes:"
	@echo "  make build           Construir imagen Docker"
	@echo "  make push            Publicar imagen"
	@echo "  make save            Exportar imagen TAR"
	@echo "  make load            Importar imagen a K3s y borrar TAR"
	@echo "  make apply           Aplicar manifiestos Kubernetes"
	@echo "  make delete          Eliminar recursos definidos en k8s/"
	@echo "  make restart         Reiniciar Deployment"
	@echo "  make redeploy        Build + Load + Apply + esperar rollout"
	@echo "  make logs            Seguir logs"
	@echo "  make previous-logs   Ver logs del contenedor anterior"
	@echo "  make status          Ver Pods"
	@echo "  make all             Ver recursos del namespace"
	@echo "  make describe        Describir Deployment"
	@echo "  make events          Ver eventos"
	@echo "  make top             Ver uso CPU/RAM"
	@echo "  make pvc             Ver PVC"
	@echo "  make shell           Abrir shell en el contenedor"
	@echo "  make images          Ver imágenes en K3s"
	@echo "  make docker-prune    Limpiar caché/imágenes Docker"
	@echo "  make clean           Borrar TAR local"
	@echo ""
	@echo "Calidad y pruebas:"
	@echo "  make test            Pruebas normales con -race"
	@echo "  make cover           Cobertura unitaria (coverage.out/html, test-report.json)"
	@echo "  make cover-all       Cobertura unitaria + integración (requiere entorno de prueba)"
	@echo "  make test-defects    Ejecutar pruebas de defectos conocidos (fallos esperados)"
	@echo "  make test-integration Ejecutar pruebas de integración (requiere entorno de prueba)"
	@echo "  make lint            Ejecutar golangci-lint"
	@echo "  make sonar           Generar cobertura y ejecutar sonar-scanner"
	@echo "  make smoke           Pruebas HTTP/REST por Kong"
	@echo "  make k6-grpc         Prueba de carga gRPC"
	@echo "  make k6-rest         Prueba de carga REST/Kong"
	@echo "  make test-up         Levantar Docker Compose de pruebas"
	@echo "  make test-down       Detener Compose de pruebas sin borrar volúmenes"
	@echo ""

# # ============================================================
# # SDATA Microservice
# # ============================================================

# SERVICE := ms-plan
# IMAGE := $(SERVICE):latest
# NAMESPACE := sdata
# TAR := $(SERVICE).tar

# KUBECTL := sudo kubectl
# CTR := sudo k3s ctr

# # ============================================================
# # Docker
# # ============================================================

# build:
# 	docker build -t $(IMAGE) .

# push:
# 	docker push $(IMAGE)

# save:
# 	docker save $(IMAGE) -o $(TAR)

# load: save
# 	$(CTR) images import $(TAR)
# 	rm -f $(TAR)

# images:
# 	$(CTR) images ls | grep $(SERVICE)

# docker-prune:
# 	docker builder prune -f
# 	docker image prune -f

# # ============================================================
# # Kubernetes
# # ============================================================

# apply:
# 	$(KUBECTL) apply -f k8s/

# delete:
# 	$(KUBECTL) delete -f k8s/

# restart:
# 	$(KUBECTL) rollout restart deployment/$(SERVICE) -n $(NAMESPACE)

# redeploy: build load apply
# 	$(KUBECTL) rollout status deployment/$(SERVICE) -n $(NAMESPACE)

# # ============================================================
# # Logs
# # ============================================================

# logs:
# 	$(KUBECTL) logs deployment/$(SERVICE) -n $(NAMESPACE) -f

# previous-logs:
# 	$(KUBECTL) logs deployment/$(SERVICE) -n $(NAMESPACE) --previous

# # ============================================================
# # Estado
# # ============================================================

# status:
# 	$(KUBECTL) get pods -n $(NAMESPACE)

# all:
# 	$(KUBECTL) get all -n $(NAMESPACE)

# describe:
# 	$(KUBECTL) describe deployment/$(SERVICE) -n $(NAMESPACE)

# events:
# 	$(KUBECTL) get events -n $(NAMESPACE) --sort-by=.metadata.creationTimestamp

# # ============================================================
# # Recursos
# # ============================================================

# top:
# 	$(KUBECTL) top pod -n $(NAMESPACE)

# pvc:
# 	$(KUBECTL) get pvc -n $(NAMESPACE)

# # ============================================================
# # Shell
# # ============================================================

# shell:
# 	$(KUBECTL) exec -it deployment/$(SERVICE) -n $(NAMESPACE) -- sh

# # ============================================================
# # Limpieza
# # ============================================================

# clean:
# 	rm -f $(TAR)

# # ============================================================
# # Ayuda
# # ============================================================

# help:
# 	@echo ""
# 	@echo "Comandos disponibles:"
# 	@echo ""
# 	@echo "  make build          -> Construir imagen Docker"
# 	@echo "  make save           -> Exportar imagen"
# 	@echo "  make load           -> Importar imagen a K3s"
# 	@echo "  make apply          -> Aplicar manifiestos"
# 	@echo "  make delete         -> Eliminar recursos"
# 	@echo "  make restart        -> Reiniciar Deployment"
# 	@echo "  make redeploy       -> Build + Load + Restart"
# 	@echo "  make logs           -> Ver logs"
# 	@echo "  make previous-logs  -> Logs del reinicio anterior"
# 	@echo "  make status         -> Ver Pods"
# 	@echo "  make all            -> Ver todos los recursos"
# 	@echo "  make top            -> Consumo CPU/RAM"
# 	@echo "  make pvc            -> Ver volúmenes"
# 	@echo "  make describe       -> Describe Deployment"
# 	@echo "  make events         -> Eventos"
# 	@echo "  make shell          -> Entrar al contenedor"
# 	@echo "  make images         -> Ver imágenes en K3s"
# 	@echo "  make clean          -> Eliminar archivo TAR"
# 	@echo "  make docker-prune   -> Limpiar imágenes y builders de Docker"
