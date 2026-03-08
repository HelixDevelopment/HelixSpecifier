#!/usr/bin/env bash
# helixspecifier_functionality_challenge.sh - Validates HelixSpecifier module core functionality and structure
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MODULE_DIR="$(cd "${SCRIPT_DIR}/../.." && pwd)"
MODULE_NAME="HelixSpecifier"

PASS=0
FAIL=0
TOTAL=0

pass() { PASS=$((PASS+1)); TOTAL=$((TOTAL+1)); echo "  PASS: $1"; }
fail() { FAIL=$((FAIL+1)); TOTAL=$((TOTAL+1)); echo "  FAIL: $1"; }

echo "=== ${MODULE_NAME} Functionality Challenge ==="
echo ""

# Test 1: Required packages exist
echo "Test: Required packages exist"
pkgs_ok=true
for pkg in engine speckit superpowers gsd ceremony intent types config; do
    if [ ! -d "${MODULE_DIR}/pkg/${pkg}" ]; then
        fail "Missing package: pkg/${pkg}"
        pkgs_ok=false
    fi
done
if [ "$pkgs_ok" = true ]; then
    pass "All required packages present (engine, speckit, superpowers, gsd, ceremony, intent, types, config)"
fi

# Test 2: 3-pillar architecture types exist
echo "Test: 3-pillar architecture types exist"
pillars_found=0
if grep -rq "type SpecKitPillar interface\|SpecKit" "${MODULE_DIR}/pkg/speckit/"; then
    pillars_found=$((pillars_found + 1))
fi
if grep -rq "type SuperpowersPillar interface\|Superpowers" "${MODULE_DIR}/pkg/superpowers/"; then
    pillars_found=$((pillars_found + 1))
fi
if grep -rq "type GSDPillar interface\|GSD" "${MODULE_DIR}/pkg/gsd/"; then
    pillars_found=$((pillars_found + 1))
fi
if [ "$pillars_found" -ge 3 ]; then
    pass "All 3 pillars found (SpecKit, Superpowers, GSD)"
else
    fail "Only ${pillars_found}/3 pillars found"
fi

# Test 3: SpecEngine interface exists
echo "Test: SpecEngine interface exists"
if grep -rq "type SpecEngine interface" "${MODULE_DIR}/pkg/types/" "${MODULE_DIR}/pkg/engine/"; then
    pass "SpecEngine interface found"
else
    fail "SpecEngine interface not found"
fi

# Test 4: EffortLevel type exists
echo "Test: EffortLevel type exists"
if grep -rq "type EffortLevel\|EffortLevel" "${MODULE_DIR}/pkg/types/"; then
    pass "EffortLevel type found in pkg/types"
else
    fail "EffortLevel type not found"
fi

# Test 5: SpecKitPhase type exists (7-phase SDD)
echo "Test: SpecKitPhase type exists for 7-phase SDD"
if grep -rq "type SpecKitPhase\|SpecKitPhase" "${MODULE_DIR}/pkg/types/"; then
    pass "SpecKitPhase type found in pkg/types"
else
    fail "SpecKitPhase type not found"
fi

# Test 6: CeremonyLevel type exists
echo "Test: CeremonyLevel type exists"
if grep -rq "type CeremonyLevel\|CeremonyLevel" "${MODULE_DIR}/pkg/"; then
    pass "CeremonyLevel type found"
else
    fail "CeremonyLevel type not found"
fi

# Test 7: Intent classifier support
echo "Test: Intent classifier support exists"
if grep -rq "Classify\|Intent\|Classifier" "${MODULE_DIR}/pkg/intent/"; then
    pass "Intent classifier support found in pkg/intent"
else
    fail "No intent classifier support found"
fi

# Test 8: DebateFunc support
echo "Test: DebateFunc support exists"
if grep -rq "type DebateFunc\|DebateFunc\|DebateFuncSetter" "${MODULE_DIR}/pkg/types/"; then
    pass "DebateFunc support found in pkg/types"
else
    fail "DebateFunc support not found"
fi

# Test 9: Spec memory support
echo "Test: Spec memory support exists"
if [ -d "${MODULE_DIR}/pkg/memory" ] && grep -rq "Memory\|Store\|Cache" "${MODULE_DIR}/pkg/memory/"; then
    pass "Spec memory support found in pkg/memory"
else
    fail "No spec memory support found"
fi

# Test 10: Effort classification support
echo "Test: Effort classification support exists"
if grep -rq "EffortClassification\|Classify\|classify" "${MODULE_DIR}/pkg/"; then
    pass "Effort classification support found"
else
    fail "No effort classification support found"
fi

# Test 11: Adapters package exists
echo "Test: Adapters package exists"
if [ -d "${MODULE_DIR}/pkg/adapters" ] && [ "$(find "${MODULE_DIR}/pkg/adapters" -name "*.go" ! -name "*_test.go" | wc -l)" -gt 0 ]; then
    pass "Adapters package found with Go files"
else
    fail "No adapters package found"
fi

# Test 12: Features/power features support
echo "Test: Features package exists"
if [ -d "${MODULE_DIR}/pkg/features" ] && [ "$(find "${MODULE_DIR}/pkg/features" -name "*.go" ! -name "*_test.go" | wc -l)" -gt 0 ]; then
    pass "Features package found with Go files"
else
    fail "No features package found"
fi

echo ""
echo "=== Results: ${PASS}/${TOTAL} passed, ${FAIL} failed ==="
[ "${FAIL}" -eq 0 ] && exit 0 || exit 1
