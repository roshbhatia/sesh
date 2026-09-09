#!/usr/bin/env bash
# shellcheck disable=SC2329 # Test functions are dispatched by name through run_test.
# Integration tests for the sy binary.
# Each test function runs the compiled binary against real filesystem state.
# Tests are self-contained: they use isolated XDG_STATE_HOME directories.

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

run_test() {
  local name="$1"
  local fn="$2"
  local tmp
  tmp=$(mktemp -d)
  export XDG_STATE_HOME="$tmp/state"
  export XDG_CONFIG_HOME="$tmp/config"
  export SYSINIT_PATHS_MANIFEST="$tmp/paths.json"

  if "$fn" "$tmp" 2>&1; then
    echo "  PASS  $name"
    PASS=$((PASS + 1))
  else
    echo "  FAIL  $name"
    FAIL=$((FAIL + 1))
    ERRORS+=("$name")
  fi
  rm -rf "$tmp"
}

assert_contains() {
  local haystack="$1" needle="$2"
  if ! echo "$haystack" | grep -qF "$needle"; then
    echo "  assertion failed: expected to find $(printf '%q' "$needle") in output"
    echo "  output was: $haystack"
    return 1
  fi
}

assert_not_contains() {
  local haystack="$1" needle="$2"
  if echo "$haystack" | grep -qF "$needle"; then
    echo "  assertion failed: expected NOT to find $(printf '%q' "$needle") in output"
    return 1
  fi
}

assert_exit_zero() {
  local code="$1" name="$2"
  if [ "$code" -ne 0 ]; then
    echo "  assertion failed: $name exited $code (expected 0)"
    return 1
  fi
}

assert_exit_nonzero() {
  local code="$1" name="$2"
  if [ "$code" -eq 0 ]; then
    echo "  assertion failed: $name exited 0 (expected non-zero)"
    return 1
  fi
}

make_git_repo() {
  local dir="$1"
  mkdir -p "$dir"
  git -C "$dir" init -q
  git -C "$dir" config user.email "t@t.com"
  git -C "$dir" config user.name "T"
  echo "hi" > "$dir/f"
  git -C "$dir" add .
  git -C "$dir" commit -m "init" -q
}

# ── tests ────────────────────────────────────────────────────────────────────

test_version() {
  local tmp="$1"
  out=$(sy --version)
  assert_contains "$out" "5.0.0"
}

test_list_empty() {
  local tmp="$1"
  out=$(sy list 2>&1 >/dev/null)
  assert_contains "$out" "No sessions"
}

test_list_alias_ls() {
  local tmp="$1"
  out=$(sy ls 2>&1 >/dev/null)
  assert_contains "$out" "No sessions"
}

test_new_invalid_name_spaces() {
  local tmp="$1"
  sy new "bad name" 2>&1 && return 1 || true
}

test_new_invalid_name_empty() {
  local tmp="$1"
  sy new "" 2>&1 && return 1 || true
}

test_path_known_session() {
  local tmp="$1"
  make_git_repo "$tmp/repo"
  # Create session directory directly (bypasses interactive picker)
  sess_root="$tmp/state/seshy/sessions"
  mkdir -p "$sess_root/my-session"
  out=$(sy path my-session)
  assert_contains "$out" "my-session"
}

test_path_unknown_session() {
  local tmp="$1"
  sy path no-such-session 2>&1 && return 1 || true
}

test_greedy_exact_match() {
  local tmp="$1"
  sess_root="$tmp/state/seshy/sessions"
  mkdir -p "$sess_root/platform-auth"
  mkdir -p "$sess_root/platform-core"
  out=$(sy --greedy platform-auth)
  assert_contains "$out" "platform-auth"
  assert_not_contains "$out" "platform-core"
}

test_greedy_prefix_match() {
  local tmp="$1"
  sess_root="$tmp/state/seshy/sessions"
  mkdir -p "$sess_root/platform-auth"
  mkdir -p "$sess_root/infra"
  out=$(sy --greedy plat)
  assert_contains "$out" "platform-auth"
}

test_greedy_substring_match() {
  local tmp="$1"
  sess_root="$tmp/state/seshy/sessions"
  mkdir -p "$sess_root/my-platform-v2"
  out=$(sy --greedy platform)
  assert_contains "$out" "my-platform-v2"
}

test_greedy_no_match_errors() {
  local tmp="$1"
  sess_root="$tmp/state/seshy/sessions"
  mkdir -p "$sess_root/alpha"
  sy --greedy zzz 2>&1 && return 1 || true
}

test_greedy_exact_beats_prefix() {
  local tmp="$1"
  sess_root="$tmp/state/seshy/sessions"
  mkdir -p "$sess_root/platform"
  mkdir -p "$sess_root/platform-extra"
  out=$(sy --greedy platform)
  # exact match should win
  trimmed=$(echo "$out" | xargs basename)
  if [ "$trimmed" != "platform" ]; then
    echo "  expected exact match 'platform', got '$trimmed'"
    return 1
  fi
}

test_delete_nonexistent_errors() {
  local tmp="$1"
  sy delete no-such 2>&1 && return 1 || true
}

test_delete_known_session() {
  local tmp="$1"
  make_git_repo "$tmp/repo"
  sess_root="$tmp/state/seshy/sessions"
  mkdir -p "$sess_root/to-delete"
  out=$(sy delete --force to-delete 2>&1)
  assert_contains "$out" "to-delete"
  if [ -d "$sess_root/to-delete" ]; then
    echo "  session directory still exists after delete"
    return 1
  fi
}

test_delete_alias_rm() {
  local tmp="$1"
  sess_root="$tmp/state/seshy/sessions"
  mkdir -p "$sess_root/rm-me"
  sy rm --force rm-me > /dev/null
  if [ -d "$sess_root/rm-me" ]; then
    echo "  session still exists after rm"
    return 1
  fi
}

test_list_shows_session() {
  local tmp="$1"
  sess_root="$tmp/state/seshy/sessions"
  mkdir -p "$sess_root/visible-session"
  out=$(sy list)
  assert_contains "$out" "visible-session"
}

test_list_shows_multiple() {
  local tmp="$1"
  sess_root="$tmp/state/seshy/sessions"
  mkdir -p "$sess_root/alpha" "$sess_root/beta" "$sess_root/gamma"
  out=$(sy list)
  assert_contains "$out" "alpha"
  assert_contains "$out" "beta"
  assert_contains "$out" "gamma"
}

test_bash_wrapper_reserves_commands() {
  local tmp="$1"
  local binary_path
  local integration
  local out
  local resolved
  local session_path="$tmp/state/seshy/sessions/list"
  mkdir -p "$session_path"
  binary_path=$(dirname "$sy_binary")
  integration=$(sy init bash)

  out=$(PATH="$binary_path:$PATH" bash --noprofile --norc -c 'eval "$1"; sy list' -- "$integration")
  assert_contains "$out" "SESSION"
  assert_contains "$out" "list"

  resolved=$(PATH="$binary_path:$PATH" bash --noprofile --norc -c 'eval "$1"; sy --greedy list' -- "$integration")
  if [[ $resolved != "$session_path" ]]; then
    echo "  expected explicit --greedy to resolve $session_path, got $resolved"
    return 1
  fi
}

test_worktree_create_and_path() {
  local tmp="$1"
  make_git_repo "$tmp/repo"
  sess_root="$tmp/state/seshy/sessions"
  mkdir -p "$sess_root/wt-session"
  # Use git worktree directly as the binary's new cmd requires interactive picker
  git -C "$tmp/repo" worktree add --detach "$sess_root/wt-session/repo-wt-session" HEAD -q
  out=$(sy path wt-session)
  assert_contains "$out" "wt-session"
}

test_list_format_json_carries_id_and_archived() {
  local tmp="$1"
  sess_root="$tmp/state/seshy/sessions"
  mkdir -p "$sess_root/plumb"
  out=$(sy list --format json)
  assert_contains "$out" '"id": "seshy:plumb"'
  assert_contains "$out" '"archived": false'
  alias_out=$(sy list --json)
  if [ "$out" != "$alias_out" ]; then
    echo "  --format json and --json differ"
    return 1
  fi
}

test_open_json_and_missing_exits_three() {
  local tmp="$1"
  sess_root="$tmp/state/seshy/sessions"
  mkdir -p "$sess_root/enter"
  out=$(sy open enter --format json)
  assert_contains "$out" '"version": "seshy.open/v1"'
  assert_contains "$out" '"SESHY_SESSION": "enter"'
  path=$(sy open enter)
  assert_contains "$path" "enter"
  sy open gone --format json 2>/dev/null && code=0 || code=$?
  [ "$code" -eq 3 ] || { echo "  open missing exited $code, want 3"; return 1; }
}

test_source_list_is_a_catalog() {
  local tmp="$1"
  sess_root="$tmp/state/seshy/sessions"
  mkdir -p "$sess_root/rowed"
  out=$(sy source list)
  assert_contains "$out" '"version": "roster.catalog/v1"'
  assert_contains "$out" '"source": "seshy"'
  assert_contains "$out" '"id": "seshy:rowed"'
  assert_contains "$out" '"kind": "local"'
}

test_provider_validate_and_list() {
  local tmp="$1"
  sess_root="$tmp/state/seshy/sessions"
  mkdir -p "$sess_root/one" "$sess_root/two"
  frame='{"version":"provider/v1","kind":"request","requestId":"r","capability":"provider.validate"}'
  out=$(printf '%s' "$frame" | sy provider)
  assert_contains "$out" '"status":"ok"'
  assert_contains "$out" '{"ok":true}'
  list_frame='{"version":"provider/v1","kind":"request","requestId":"r2","capability":"source.list"}'
  list_out=$(printf '%s' "$list_frame" | sy provider)
  assert_contains "$list_out" 'roster.catalog/v1'
  assert_contains "$list_out" 'seshy:one'
  assert_contains "$list_out" 'seshy:two'
}

test_delete_yes_skips_prompt() {
  local tmp="$1"
  sess_root="$tmp/state/seshy/sessions"
  mkdir -p "$sess_root/drop"
  sy delete --yes drop </dev/null >/dev/null 2>&1
  if [ -d "$sess_root/drop" ]; then
    echo "  --yes did not delete the session"
    return 1
  fi
  mkdir -p "$sess_root/refuse"
  sy delete refuse </dev/null 2>/dev/null && code=0 || code=$?
  [ "$code" -eq 4 ] || { echo "  delete without a terminal exited $code, want 4"; return 1; }
  [ -d "$sess_root/refuse" ] || { echo "  refused delete removed the session"; return 1; }
}

test_prune_dry_run_lists_orphan_branch() {
  local tmp="$1"
  make_git_repo "$tmp/repo"
  sy new feat "$tmp/repo" >/dev/null 2>&1
  sess_root="$tmp/state/seshy/sessions"
  rm -rf "$sess_root/feat"
  out=$(sy prune --dry-run "$tmp/repo" 2>&1)
  assert_contains "$out" "would delete branch sy/feat/repo"
  git -C "$tmp/repo" rev-parse --verify --quiet refs/heads/sy/feat/repo >/dev/null || {
    echo "  dry run deleted the branch"
    return 1
  }
}

test_git_config_branch_format_override() {
  local tmp="$1"
  make_git_repo "$tmp/repo"
  git -C "$tmp/repo" config seshy.branchFormat 'gitcfg/{{.Session}}/{{.Repo}}'
  sy new feat "$tmp/repo" >/dev/null 2>&1
  git -C "$tmp/repo" rev-parse --verify --quiet refs/heads/gitcfg/feat/repo >/dev/null || {
    echo "  git config seshy.branchFormat did not name the branch"
    return 1
  }
}

# ── run all ──────────────────────────────────────────────────────────────────

echo "Running integration tests..."
echo ""

run_test "version flag" test_version
run_test "list empty" test_list_empty
run_test "list alias ls" test_list_alias_ls
run_test "new invalid name (spaces)" test_new_invalid_name_spaces
run_test "new invalid name (empty)" test_new_invalid_name_empty
run_test "path known session" test_path_known_session
run_test "path unknown session errors" test_path_unknown_session
run_test "greedy exact match" test_greedy_exact_match
run_test "greedy prefix match" test_greedy_prefix_match
run_test "greedy substring match" test_greedy_substring_match
run_test "greedy no match errors" test_greedy_no_match_errors
run_test "greedy exact beats prefix" test_greedy_exact_beats_prefix
run_test "delete nonexistent errors" test_delete_nonexistent_errors
run_test "delete known session" test_delete_known_session
run_test "delete alias rm" test_delete_alias_rm
run_test "list shows session" test_list_shows_session
run_test "list shows multiple" test_list_shows_multiple
run_test "shell wrapper reserves commands" test_bash_wrapper_reserves_commands
run_test "worktree create and path" test_worktree_create_and_path
run_test "list --format json carries id and archived" test_list_format_json_carries_id_and_archived
run_test "open json and missing exits 3" test_open_json_and_missing_exits_three
run_test "source list is a catalog" test_source_list_is_a_catalog
run_test "provider validate and list" test_provider_validate_and_list
run_test "delete --yes skips the prompt" test_delete_yes_skips_prompt
run_test "prune --dry-run lists the orphan branch" test_prune_dry_run_lists_orphan_branch
run_test "git config branchFormat override" test_git_config_branch_format_override

echo ""
echo "Results: $PASS passed, $FAIL failed"

if [ "${#ERRORS[@]}" -gt 0 ]; then
  echo "Failed tests:"
  for e in "${ERRORS[@]}"; do
    echo "  - $e"
  done
  exit 1
fi

exit 0
