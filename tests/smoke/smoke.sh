#!/usr/bin/env bash
set -euo pipefail

BASE="${1:-http://localhost:8080}/api/v1"

echo "=== Smoke Tests ==="
echo "Target: $BASE"
echo ""

# Health
echo "--- Liveness ---"
curl -sf "${BASE%/api/v1}/healthz" | python3 -c "import sys,json; d=json.load(sys.stdin); assert d['status']=='ok'" && echo "PASS"

echo "--- Readiness ---"
curl -sf "${BASE%/api/v1}/readyz" | python3 -c "import sys,json; d=json.load(sys.stdin); assert d['status']=='ok'" && echo "PASS"

# Reload defs
echo "--- Reload Defs ---"
RELOAD=$(curl -sf -X POST "$BASE/admin/reload-defs")
echo "$RELOAD" | python3 -c "import sys,json; d=json.load(sys.stdin); assert 'loaded' in d" && echo "PASS"

# List defs
echo "--- List Defs ---"
DEFS=$(curl -sf "$BASE/workflow-defs")
DEF_COUNT=$(echo "$DEFS" | python3 -c "import sys,json; print(json.load(sys.stdin)['total'])")
echo "  Found $DEF_COUNT definitions"
if [ "$DEF_COUNT" -eq "0" ]; then
  echo "SKIP: No workflow defs found. Add JSON files to workflows/ directory."
  exit 0
fi

DEF_ID=$(echo "$DEFS" | python3 -c "
import sys, json
items = json.load(sys.stdin)['items']
linear = next((d for d in items if d['name'] == 'leave-approval'), None)
print((linear or items[0])['id'])
")
echo "  Using def: $DEF_ID"

# Start instance
echo "--- Start Instance ---"
INST=$(curl -sf -X POST "$BASE/workflow-instances" \
  -H "Content-Type: application/json" \
  -d "{\"workflow_def_id\":\"$DEF_ID\",\"created_by\":\"smoke-test\"}")
INST_ID=$(echo "$INST" | python3 -c "import sys,json; print(json.load(sys.stdin)['id'])")
STATUS=$(echo "$INST" | python3 -c "import sys,json; print(json.load(sys.stdin)['status'])")
[ "$STATUS" = "active" ] && echo "PASS (id: ${INST_ID:0:8}...)" || (echo "FAIL: expected active, got $STATUS"; exit 1)

# Get instance
echo "--- Get Instance ---"
curl -sf "$BASE/workflow-instances/$INST_ID" | python3 -c "import sys,json; d=json.load(sys.stdin); assert d['id']==\"$INST_ID\"" && echo "PASS"

# List tasks
echo "--- List Tasks ---"
TASKS=$(curl -sf "$BASE/workflow-instances/$INST_ID/tasks")
TASK_COUNT=$(echo "$TASKS" | python3 -c "import sys,json; print(len(json.load(sys.stdin)))")
echo "  Found $TASK_COUNT tasks"
if [ "$TASK_COUNT" -eq "0" ]; then
  echo "SKIP: Start state has no tasks."
  exit 0
fi
TASK_ID=$(echo "$TASKS" | python3 -c "import sys,json; print(json.load(sys.stdin)[0]['id'])")

# Get task
echo "--- Get Task ---"
curl -sf "$BASE/task-instances/$TASK_ID" | python3 -c "import sys,json; d=json.load(sys.stdin); assert d['id']==\"$TASK_ID\"" && echo "PASS"

# Complete task
echo "--- Complete Task ---"
RESULT=$(curl -sf -X POST "$BASE/task-instances/$TASK_ID/complete" \
  -H "Content-Type: application/json" \
  -d '{"status":"done","completed_by":"smoke-test"}')
echo "$RESULT" | python3 -m json.tool > /dev/null && echo "PASS"

# History
echo "--- History ---"
HIST=$(curl -sf "$BASE/workflow-instances/$INST_ID/history")
HIST_COUNT=$(echo "$HIST" | python3 -c "import sys,json; print(len(json.load(sys.stdin)))")
echo "  Transitions: $HIST_COUNT"

echo ""
echo "=== All smoke tests passed ==="
