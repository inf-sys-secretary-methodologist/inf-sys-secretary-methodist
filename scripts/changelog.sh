#!/usr/bin/env bash
# ------------------------------------------------------------------
# changelog.sh — инкрементально добавляет секцию новой версии в CHANGELOG.md
# через git-cliff (конфиг cliff.toml в корне).
#
# Старая детальная (ручная) история СОХРАНЯЕТСЯ: git-cliff берёт только
# коммиты после последнего релизного тега (--unreleased) и генерит один
# СКЕЛЕТ-секции, который вставляется сразу под шапкой, перед самой свежей
# версией. Скелет затем дополняется вручную (детали слайсов, номера задач).
#
# Использование:
#   scripts/changelog.sh --preview      показать, что попадёт в след. версию
#   scripts/changelog.sh v0.223.0       добавить секцию для нового тега
#
# Место в релизном процессе (проект работает через PR, не прямой push в main):
#   1. scripts/changelog.sh vX.Y.Z      # скелет секции в CHANGELOG.md
#   2. дополнить секцию руками (детали, #issue, TDD-заметки)
#   3. _tools/bump_version.sh X.Y.Z    # синхронизировать 8 файлов версии
#   4. ветка chore/vX.Y.Z-desc -> commit -> PR -> squash-merge в main
#   5. git tag vX.Y.Z на смердженном коммите + gh release
# ------------------------------------------------------------------
set -euo pipefail

# Корень репозитория из расположения скрипта, не из cwd вызывающего.
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_ROOT"

# git-cliff из PATH, иначе разово через nix (без глобальной установки).
if command -v git-cliff >/dev/null 2>&1; then
  CLIFF=(git-cliff)
elif command -v nix >/dev/null 2>&1; then
  CLIFF=(nix run nixpkgs#git-cliff --)
else
  echo "нужен git-cliff (или nix). Добавь pkgs.git-cliff в nix-конфиг." >&2
  exit 1
fi

CFG=cliff.toml
FILE=CHANGELOG.md

if [ "${1:-}" = "--preview" ]; then
  "${CLIFF[@]}" --config "$CFG" --unreleased --strip all
  exit 0
fi

TAG="${1:?укажи тег новой версии, напр. v0.223.0}"

TMPNEW="$(mktemp)"
trap 'rm -f "$TMPNEW"' EXIT
"${CLIFF[@]}" --config "$CFG" --unreleased --tag "$TAG" --strip all > "$TMPNEW"

if [ -z "$(tr -d '[:space:]' < "$TMPNEW")" ]; then
  echo "нет новых conventional-коммитов после последнего тега - CHANGELOG не изменён." >&2
  exit 0
fi

# Вставляем секцию перед первой существующей версией (## [...]), т.е. сразу
# под шапкой. Секцию читаем из файла (getline) - переносимо на macOS awk.
awk -v f="$TMPNEW" '
  !ins && /^## \[/ {
    while ((getline line < f) > 0) print line
    print ""
    ins=1
  }
  { print }
  END { if (!ins) { print ""; while ((getline line < f) > 0) print line } }
' "$FILE" > "$FILE.tmp" && mv "$FILE.tmp" "$FILE"

echo "CHANGELOG.md обновлён: добавлен скелет секции $TAG — дополни детали руками."
