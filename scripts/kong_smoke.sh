#!/usr/bin/env bash
# Pruebas de humo REST contra Kong -> ms-plan (transcodificacion REST -> gRPC).
# Equivale a los casos de prueba que se armarian en SoapUI (ver docs/PLAN_DE_PRUEBAS.md).
#
# Uso:  BASE_URL=https://192.168.0.240 ./scripts/kong_smoke.sh
# Requiere: curl y jq. Imprime un resumen PASS/FAIL y sale con codigo != 0 si algo falla.

set -u
BASE="${BASE_URL:-https://192.168.0.240}"
CURL=(curl -sk -o /tmp/smoke_body.json -w '%{http_code}|%{time_total}' -H 'Content-Type: application/json')

PASS=0
FAIL=0
TITLE="SMOKE-$(date +%s)"
ID=""

check() { # nombre esperado obtenido tiempo
  if [[ "$2" == "$3" ]]; then
    printf 'PASS  %-46s http=%s  t=%ss\n' "$1" "$3" "$4"; PASS=$((PASS + 1))
  else
    printf 'FAIL  %-46s esperado=%s obtenido=%s  t=%ss\n' "$1" "$2" "$3" "$4"; FAIL=$((FAIL + 1))
  fi
}

run() { # nombre esperado metodo url [cuerpo]
  local name="$1" want="$2" method="$3" url="$4" body="${5:-}" out code t
  if [[ -n "$body" ]]; then
    out=$("${CURL[@]}" -X "$method" "$url" -d "$body")
  else
    out=$("${CURL[@]}" -X "$method" "$url")
  fi
  code="${out%%|*}"; t="${out##*|}"
  check "$name" "$want" "$code" "$t"
}

PLAN='{"data":{"plan_title":"'"$TITLE"'","plan_subtitle":"humo","plan_type":"monthly","plan_price":99900,"active_plan":true}}'

run "TC-01 POST /plan crea plan"                200 POST "$BASE/api/v1/plan" "$PLAN"
ID=$(jq -r '.data.id // empty' /tmp/smoke_body.json)
[[ -n "$ID" ]] || { echo "No se obtuvo id; revise nombres de campos del JSON"; cat /tmp/smoke_body.json; exit 2; }

run "TC-02 POST /plan titulo duplicado"         409 POST "$BASE/api/v1/plan" "$PLAN"
run "TC-03 POST /plan sin titulo"               400 POST "$BASE/api/v1/plan" '{"data":{"plan_subtitle":"x"}}'
run "TC-04 GET /plan/{id} existente"            200 GET  "$BASE/api/v1/plan/$ID"
run "TC-05 GET /plan/{id} inexistente"          404 GET  "$BASE/api/v1/plan/000000000000000000000001"
run "TC-06 GET /plan/{id} id invalido"          400 GET  "$BASE/api/v1/plan/no-es-objectid"
run "TC-07 GET /plans lista paginada"           200 GET  "$BASE/api/v1/plans?page=1&limit=5"
run "TC-08 PUT /plan/{id} actualiza"            200 PUT  "$BASE/api/v1/plan/$ID" \
  '{"data":{"plan_title":"'"$TITLE"'","plan_subtitle":"actualizado","plan_price":120000}}'
run "TC-09 PUT /plan/{id} inexistente"          404 PUT  "$BASE/api/v1/plan/000000000000000000000001" "$PLAN"
run "TC-10 GET /plan/company/{id} (D-04)"       200 GET  "$BASE/api/v1/plan/company/000000000000000000000001"
run "TC-11 DELETE /plan/{id}/{plan_id} feature" 200 DELETE "$BASE/api/v1/plan/1/$ID"
run "TC-12 DELETE /plan/{id}"                   200 DELETE "$BASE/api/v1/plan/$ID"
run "TC-13 DELETE /plan/{id} ya borrado"        404 DELETE "$BASE/api/v1/plan/$ID"

echo "-----------------------------------------"
echo "PASS=$PASS  FAIL=$FAIL"
# TC-10 falla hoy por el defecto D-04 (GetPlanCompany no implementado): es esperado.
[[ "$FAIL" -eq 0 ]]
