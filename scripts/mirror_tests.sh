#!/usr/bin/env bash
# Zrcadlí testy z lessons/lesson-NN/exercise do lessons/lesson-NN/solutions.
#
# Test se píše jen jednou — v exercise/ jako externí testovací balíček
# (package exercise_test importující ".../exercise"). Tenhle skript z něj
# vyrobí identickou variantu pro solutions/, takže referenční řešení je
# ověřované úplně stejnými testy jako cvičení.
#
# Bez argumentů zpracuje všechny lekce. S argumenty jen vyjmenované:
#   ./scripts/mirror_tests.sh 04 05 06
set -euo pipefail

cd "$(dirname "$0")/.."

if [ "$#" -gt 0 ]; then
	dirs=()
	for arg in "$@"; do
		dirs+=("lessons/lesson-$(printf '%02d' "$((10#${arg#lesson-}))")/")
	done
else
	dirs=(lessons/lesson-*/)
fi

count=0
for dir in "${dirs[@]}"; do
	lesson="$(basename "${dir%/}")"
	src_dir="$dir/exercise"
	dst_dir="$dir/solutions"

	[ -d "$src_dir" ] || continue
	[ -d "$dst_dir" ] || continue

	# smaž staré zrcadlené testy, ať po přejmenování nezůstane sirotek
	find "$dst_dir" -name '*_test.go' -delete

	while IFS= read -r -d '' src; do
		rel="${src#"$src_dir"/}"
		dst="$dst_dir/$rel"
		mkdir -p "$(dirname "$dst")"
		sed \
			-e 's|^package exercise_test$|package solutions_test|' \
			-e 's|^package exercise$|package solutions|' \
			-e "s|/${lesson}/exercise\"|/${lesson}/solutions\"|" \
			"$src" >"$dst"
		count=$((count + 1))
	done < <(find "$src_dir" -name '*_test.go' -print0)
done

echo "zrcadleno $count testovacích souborů"
