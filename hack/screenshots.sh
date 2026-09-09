#!/usr/bin/env bash
set -euo pipefail

repo_dir=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
cd "$repo_dir"
output_dir=${SESHY_MEDIA_OUTPUT_DIR:-"$repo_dir/docs"}
mkdir -p "$output_dir"

media_fingerprint() {
  {
    printf '%s\n' flake.lock flake.nix go.mod go.sum hack/seshy.tape hack/screenshots.sh
    find cmd internal -type f -name '*.go' ! -name '*_test.go' -print | LC_ALL=C sort
  } | while IFS= read -r path; do
    sha256sum "$path"
  done | sha256sum | cut -d ' ' -f 1
}

media_is_valid() {
  [[ -s $output_dir/seshy.png && -s $output_dir/seshy.gif ]] || return 1
  [[ $(identify -format '%m' "$output_dir/seshy.png") == PNG ]]
  [[ $(identify -format '%m' "$output_dir/seshy.gif[0]") == GIF ]]
}

if [[ ${1:-} == "--check" ]]; then
  expected=$(media_fingerprint)
  current=$(cat "$output_dir/.seshy-media.sha256" 2> /dev/null || true)
  if [[ $current != "$expected" ]] || ! media_is_valid; then
    echo "Seshy media is stale; run ./hack/screenshots.sh" >&2
    exit 1
  fi
  exit 0
fi

media_root=$(mktemp -d)
trap 'rm -rf "$media_root"' EXIT
go_mod_cache=$(go env GOMODCACHE)
export HOME="$media_root/home"
export GOMODCACHE="$go_mod_cache"
export XDG_CONFIG_HOME="$media_root/config"
export XDG_STATE_HOME="$media_root/state"
mkdir -p "$HOME/src" "$XDG_CONFIG_HOME" "$XDG_STATE_HOME"
HOME=$(cd "$HOME" && pwd -P)
export HOME
export SESHY_SESSIONS_DIR="$HOME/sessions"
export SESHY_ARCHIVE_DIR="$HOME/archive"

clone_source() {
  local repository="$1" revision="$2"
  git clone --quiet "https://github.com/roshbhatia/$repository.git" "$HOME/src/$repository"
  git -C "$HOME/src/$repository" checkout --quiet --detach "$revision"
}
clone_source changes 72449fa57f2813300968e042952126f9fc32e045
clone_source ask d6bd2d4cd84aad5d677b33f229bb35124e9ca725

mkdir -p "$media_root/bin"
go build -o "$media_root/bin/sy" ./cmd/sy
export PATH="$media_root/bin:$PATH"
sy new provider-release "$HOME/src/changes" "$HOME/src/ask" > /dev/null

freeze --execute "sy status provider-release" \
  --output "$output_dir/seshy.png" \
  --width 1000 \
  --padding 24 \
  --margin 16 \
  --window

vhs "$repo_dir/hack/seshy.tape" --output "$output_dir/seshy.gif"
chmod 0644 "$output_dir/seshy.png" "$output_dir/seshy.gif"

if ! media_is_valid; then
  echo "Seshy media generation produced an invalid image" >&2
  exit 1
fi
media_fingerprint > "$output_dir/.seshy-media.sha256"
