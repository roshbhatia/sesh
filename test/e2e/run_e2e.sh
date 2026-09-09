#!/usr/bin/env bash
# shellcheck disable=SC2329 # Scenarios are dispatched by name through run_scenario.
# End-to-end tests for sy.
# Each scenario exercises a complete user workflow from start to finish,
# using only the sy binary and standard git tooling.

set -euo pipefail

repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)
test_binary_dir=""
sy_binary=""

cleanup_test_binary() {
  if [[ -n ${test_binary_dir} ]]; then
    rm -rf "${test_binary_dir:?}"
  fi
}
trap cleanup_test_binary EXIT

if [[ -n ${SESHY_TEST_BINARY:-} ]]; then
  sy_binary="${SESHY_TEST_BINARY}"
elif [[ -f "${repo_root}/go.mod" ]]; then
  test_binary_dir=$(mktemp -d)
  (cd "${repo_root}" && go build -o "${test_binary_dir}/sy" ./cmd/sy)
  sy_binary="${test_binary_dir}/sy"
elif command -v sy > /dev/null; then
  sy_binary=$(command -v sy)
else
  echo "ERROR: sy is not installed and no source checkout is available" >&2
  exit 1
fi

if [[ ! -x ${sy_binary} ]]; then
  echo "ERROR: test binary is not executable: ${sy_binary}" >&2
  exit 1
fi

sy() {
  "${sy_binary}" "$@"
}

PASS=0
FAIL=0
ERRORS=()

# ── harness ──────────────────────────────────────────────────────────────────

run_scenario() {
  local name="$1"
  local fn="$2"
  local tmp
  tmp=$(mktemp -d)
  export XDG_STATE_HOME="$tmp/state"
  export XDG_CONFIG_HOME="$tmp/config"
  export SYSINIT_PATHS_MANIFEST="$tmp/paths.json"
  export HOME="$tmp/home"
  mkdir -p "$HOME"

  echo "── $name"
  if "$fn" "$tmp" 2>&1; then
    echo "   PASS"
    PASS=$((PASS + 1))
  else
    echo "   FAIL"
    FAIL=$((FAIL + 1))
    ERRORS+=("$name")
  fi
  rm -rf "$tmp"
  echo ""
}

die() {
  echo "  ERROR: $*" >&2
  return 1
}

assert() {
  local desc="$1"
  shift
  if ! "$@"; then
    die "assertion failed: $desc"
  fi
}

assert_contains() {
  local desc="$1" haystack="$2" needle="$3"
  if ! echo "$haystack" | grep -qF "$needle"; then
    die "$desc: expected $(printf '%q' "$needle") in $(printf '%q' "$haystack")"
  fi
}

assert_not_contains() {
  local desc="$1" haystack="$2" needle="$3"
  if echo "$haystack" | grep -qF "$needle"; then
    die "$desc: expected NOT to find $(printf '%q' "$needle")"
  fi
}

assert_dir_exists() {
  local desc="$1" path="$2"
  [ -d "$path" ] || die "$desc: directory $path does not exist"
}

assert_dir_missing() {
  local desc="$1" path="$2"
  [ ! -d "$path" ] || die "$desc: directory $path should not exist"
}

assert_symlink() {
  local desc="$1" path="$2"
  [ -L "$path" ] || die "$desc: $path is not a symlink"
}

make_git_repo() {
  local dir="$1" name="${2:-repo}"
  mkdir -p "$dir"
  git -C "$dir" init -q
  git -C "$dir" config user.email "t@t.com"
  git -C "$dir" config user.name "Tester"
  echo "# $name" > "$dir/README.md"
  git -C "$dir" add .
  git -C "$dir" commit -m "init" -q
}

# ── scenarios ─────────────────────────────────────────────────────────────────

# Scenario 1: basic lifecycle — create session dir, list, path, delete
scenario_basic_lifecycle() {
  local tmp="$1"
  local sess_root="$tmp/state/seshy/sessions"

  # Start with no sessions. The advice goes to stderr; stdout stays empty for
  # consumers that parse the table.
  out=$(sy list 2>&1 >/dev/null)
  assert_contains "empty list" "$out" "No sessions"
  out=$(sy list 2>/dev/null)
  [ -z "$out" ] || die "empty list must print nothing on stdout, got $(printf '%q' "$out")"

  # Manually create a session dir (bypasses interactive picker)
  mkdir -p "$sess_root/my-feature"

  # List shows it
  out=$(sy list)
  assert_contains "list after create" "$out" "my-feature"

  # Path resolves correctly
  out=$(sy path my-feature)
  assert_contains "path output" "$out" "my-feature"

  # cd workflow: path returns a navigable directory
  cd_target=$(sy path my-feature | tr -d '\n')
  assert_dir_exists "session dir" "$cd_target"

  # Delete removes it
  sy delete --force my-feature > /dev/null
  assert_dir_missing "after delete" "$sess_root/my-feature"

  # List is empty again
  out=$(sy list 2>&1 >/dev/null)
  assert_contains "empty after delete" "$out" "No sessions"
}

# Scenario 2: git worktree workflow
scenario_git_worktree() {
  local tmp="$1"
  local sess_root="$tmp/state/seshy/sessions"

  make_git_repo "$tmp/repos/composition-runtime" "composition-runtime"
  make_git_repo "$tmp/repos/provider-metadata" "provider-metadata"

  # Create session dir and worktrees manually (picker is interactive)
  mkdir -p "$sess_root/platform-v2"
  git -C "$tmp/repos/composition-runtime" worktree add \
    --detach "$sess_root/platform-v2/composition-runtime-platform-v2" HEAD -q
  git -C "$tmp/repos/provider-metadata" worktree add \
    --detach "$sess_root/platform-v2/provider-metadata-platform-v2" HEAD -q

  # Both worktrees must be valid git repos
  assert_dir_exists "wt1" "$sess_root/platform-v2/composition-runtime-platform-v2"
  assert_dir_exists "wt2" "$sess_root/platform-v2/provider-metadata-platform-v2"
  git -C "$sess_root/platform-v2/composition-runtime-platform-v2" status -s > /dev/null ||
    die "worktree 1 is not a valid git repo"
  git -C "$sess_root/platform-v2/provider-metadata-platform-v2" status -s > /dev/null ||
    die "worktree 2 is not a valid git repo"

  # List shows repo count = 2
  out=$(sy list)
  assert_contains "repo count" "$out" "2"

  # sy path outputs the session directory
  path_out=$(sy path platform-v2 | tr -d '\n')
  assert_contains "path contains session name" "$path_out" "platform-v2"

  # sy --greedy works
  greedy_out=$(sy --greedy platform | tr -d '\n')
  assert_contains "greedy matches" "$greedy_out" "platform-v2"

  # Delete cleans up session directory including worktrees
  sy delete --force platform-v2 > /dev/null
  assert_dir_missing "session gone" "$sess_root/platform-v2"

  # Worktrees should be pruned from git's perspective
  git -C "$tmp/repos/composition-runtime" worktree prune
  wt_list=$(git -C "$tmp/repos/composition-runtime" worktree list)
  assert_not_contains "wt pruned" "$wt_list" "platform-v2"
}

# Scenario 3: symlink workflow for non-git directories
scenario_symlink_non_git() {
  local tmp="$1"
  local sess_root="$tmp/state/seshy/sessions"

  mkdir -p "$tmp/workdir/my-docs"
  echo "notes" > "$tmp/workdir/my-docs/notes.txt"

  mkdir -p "$sess_root/docs-session"
  ln -s "$tmp/workdir/my-docs" "$sess_root/docs-session/my-docs"

  assert_symlink "symlink exists" "$sess_root/docs-session/my-docs"
  assert_dir_exists "symlink resolves" "$sess_root/docs-session/my-docs"

  out=$(sy list)
  assert_contains "list shows session" "$out" "docs-session"

  sy delete --force docs-session > /dev/null
  assert_dir_missing "session gone" "$sess_root/docs-session"
}

# Scenario 4: greedy matching priority
scenario_greedy_priority() {
  local tmp="$1"
  local sess_root="$tmp/state/seshy/sessions"

  mkdir -p "$sess_root/platform" \
    "$sess_root/platform-auth" \
    "$sess_root/my-platform-v2"

  # Exact match wins
  out=$(sy --greedy platform | xargs basename)
  [ "$out" = "platform" ] || die "exact should win; got $out"

  # Prefix match wins over substring
  out=$(sy --greedy platform-a | xargs basename)
  [ "$out" = "platform-auth" ] || die "prefix should win; got $out"

  # Substring match
  sy delete --force platform > /dev/null
  sy delete --force platform-auth > /dev/null
  out=$(sy --greedy platform | xargs basename)
  [ "$out" = "my-platform-v2" ] || die "substring fallback; got $out"

  # No match → non-zero exit
  sy --greedy zzz 2> /dev/null && die "expected error for no match" || true
}

# Scenario 5: multi-session isolation
scenario_multi_session_isolation() {
  local tmp="$1"
  local sess_root="$tmp/state/seshy/sessions"

  make_git_repo "$tmp/shared-repo" "shared"

  mkdir -p "$sess_root/session-a"
  mkdir -p "$sess_root/session-b"

  git -C "$tmp/shared-repo" worktree add \
    --detach "$sess_root/session-a/shared-repo-session-a" HEAD -q
  git -C "$tmp/shared-repo" worktree add \
    --detach "$sess_root/session-b/shared-repo-session-b" HEAD -q

  # Both sessions listed
  out=$(sy list)
  assert_contains "session-a listed" "$out" "session-a"
  assert_contains "session-b listed" "$out" "session-b"

  # Each has its own path
  path_a=$(sy path session-a | tr -d '\n')
  path_b=$(sy path session-b | tr -d '\n')
  [ "$path_a" != "$path_b" ] || die "sessions should have different paths"

  # Deleting one does not affect the other
  sy delete --force session-a > /dev/null
  assert_dir_missing "session-a gone" "$sess_root/session-a"
  assert_dir_exists "session-b intact" "$sess_root/session-b"

  out=$(sy list)
  assert_not_contains "session-a gone from list" "$out" "session-a"
  assert_contains "session-b still listed" "$out" "session-b"
}

# Scenario 6: cd $(sy --greedy ...) produces single-line output
scenario_greedy_single_line_output() {
  local tmp="$1"
  local sess_root="$tmp/state/seshy/sessions"

  mkdir -p "$sess_root/nav-target"

  out=$(sy --greedy nav-target)
  lines=$(echo "$out" | wc -l | tr -d ' ')
  [ "$lines" -eq 1 ] || die "expected 1 line of output, got $lines"
  assert_contains "output contains session" "$out" "nav-target"
}

# Scenario 7: delete alias and repository removal
scenario_delete_and_remove() {
  local tmp="$1"
  local sess_root="$tmp/state/seshy/sessions"

  mkdir -p "$sess_root/rm-test"
  sy rm --force rm-test > /dev/null
  assert_dir_missing "rm alias works" "$sess_root/rm-test"

  mkdir -p "$sess_root/remove-test/api"
  sy remove --force remove-test api > /dev/null
  assert_dir_missing "repository removal works" "$sess_root/remove-test/api"
  assert_dir_exists "repository removal keeps the session" "$sess_root/remove-test"
}

# ── run all scenarios ─────────────────────────────────────────────────────────

echo "Running e2e scenarios..."
echo ""

run_scenario "basic lifecycle" scenario_basic_lifecycle
run_scenario "git worktree workflow" scenario_git_worktree
run_scenario "symlink non-git workflow" scenario_symlink_non_git
run_scenario "greedy matching priority" scenario_greedy_priority
run_scenario "multi-session isolation" scenario_multi_session_isolation
run_scenario "greedy single-line output" scenario_greedy_single_line_output
run_scenario "delete alias and repository removal" scenario_delete_and_remove

echo "Results: $PASS passed, $FAIL failed"

if [ "${#ERRORS[@]}" -gt 0 ]; then
  echo "Failed scenarios:"
  for e in "${ERRORS[@]}"; do
    echo "  - $e"
  done
  exit 1
fi

exit 0
