#!/usr/bin/env bash
# Integration tests for the sesh binary.
# Each test function runs the compiled binary against real filesystem state.
# Tests are self-contained: they use isolated XDG_STATE_HOME directories.

set -euo pipefail

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
  out=$(sesh --version)
  assert_contains "$out" "3.0.0"
}

test_list_empty() {
  local tmp="$1"
  out=$(sesh list)
  assert_contains "$out" "No sessions"
}

test_list_alias_ls() {
  local tmp="$1"
  out=$(sesh ls)
  assert_contains "$out" "No sessions"
}

test_new_invalid_name_spaces() {
  local tmp="$1"
  sesh new "bad name" 2>&1 && return 1 || true
}

test_new_invalid_name_empty() {
  local tmp="$1"
  sesh new "" 2>&1 && return 1 || true
}

test_path_known_session() {
  local tmp="$1"
  make_git_repo "$tmp/repo"
  # Create session directory directly (bypasses interactive picker)
  sess_root="$tmp/state/sesh/sessions"
  mkdir -p "$sess_root/my-session"
  out=$(sesh path my-session)
  assert_contains "$out" "my-session"
}

test_path_unknown_session() {
  local tmp="$1"
  sesh path no-such-session 2>&1 && return 1 || true
}

test_greedy_exact_match() {
  local tmp="$1"
  sess_root="$tmp/state/sesh/sessions"
  mkdir -p "$sess_root/platform-auth"
  mkdir -p "$sess_root/platform-core"
  out=$(sesh --greedy platform-auth)
  assert_contains "$out" "platform-auth"
  assert_not_contains "$out" "platform-core"
}

test_greedy_prefix_match() {
  local tmp="$1"
  sess_root="$tmp/state/sesh/sessions"
  mkdir -p "$sess_root/platform-auth"
  mkdir -p "$sess_root/infra"
  out=$(sesh --greedy plat)
  assert_contains "$out" "platform-auth"
}

test_greedy_substring_match() {
  local tmp="$1"
  sess_root="$tmp/state/sesh/sessions"
  mkdir -p "$sess_root/my-platform-v2"
  out=$(sesh --greedy platform)
  assert_contains "$out" "my-platform-v2"
}

test_greedy_no_match_errors() {
  local tmp="$1"
  sess_root="$tmp/state/sesh/sessions"
  mkdir -p "$sess_root/alpha"
  sesh --greedy zzz 2>&1 && return 1 || true
}

test_greedy_exact_beats_prefix() {
  local tmp="$1"
  sess_root="$tmp/state/sesh/sessions"
  mkdir -p "$sess_root/platform"
  mkdir -p "$sess_root/platform-extra"
  out=$(sesh --greedy platform)
  # exact match should win
  trimmed=$(echo "$out" | xargs basename)
  if [ "$trimmed" != "platform" ]; then
    echo "  expected exact match 'platform', got '$trimmed'"
    return 1
  fi
}

test_delete_nonexistent_errors() {
  local tmp="$1"
  sesh delete no-such 2>&1 && return 1 || true
}

test_delete_known_session() {
  local tmp="$1"
  make_git_repo "$tmp/repo"
  sess_root="$tmp/state/sesh/sessions"
  mkdir -p "$sess_root/to-delete"
  out=$(sesh delete to-delete)
  assert_contains "$out" "to-delete"
  if [ -d "$sess_root/to-delete" ]; then
    echo "  session directory still exists after delete"
    return 1
  fi
}

test_delete_alias_rm() {
  local tmp="$1"
  sess_root="$tmp/state/sesh/sessions"
  mkdir -p "$sess_root/rm-me"
  sesh rm rm-me > /dev/null
  if [ -d "$sess_root/rm-me" ]; then
    echo "  session still exists after rm"
    return 1
  fi
}

test_list_shows_session() {
  local tmp="$1"
  sess_root="$tmp/state/sesh/sessions"
  mkdir -p "$sess_root/visible-session"
  out=$(sesh list)
  assert_contains "$out" "visible-session"
}

test_list_shows_multiple() {
  local tmp="$1"
  sess_root="$tmp/state/sesh/sessions"
  mkdir -p "$sess_root/alpha" "$sess_root/beta" "$sess_root/gamma"
  out=$(sesh list)
  assert_contains "$out" "alpha"
  assert_contains "$out" "beta"
  assert_contains "$out" "gamma"
}

test_worktree_create_and_path() {
  local tmp="$1"
  make_git_repo "$tmp/repo"
  sess_root="$tmp/state/sesh/sessions"
  mkdir -p "$sess_root/wt-session"
  # Use git worktree directly as the binary's new cmd requires interactive picker
  git -C "$tmp/repo" worktree add --detach "$sess_root/wt-session/repo-wt-session" HEAD -q
  out=$(sesh path wt-session)
  assert_contains "$out" "wt-session"
}

# ── run all ──────────────────────────────────────────────────────────────────

echo "Running integration tests..."
echo ""

run_test "version flag"                  test_version
run_test "list empty"                    test_list_empty
run_test "list alias ls"                 test_list_alias_ls
run_test "new invalid name (spaces)"     test_new_invalid_name_spaces
run_test "new invalid name (empty)"      test_new_invalid_name_empty
run_test "path known session"            test_path_known_session
run_test "path unknown session errors"   test_path_unknown_session
run_test "greedy exact match"            test_greedy_exact_match
run_test "greedy prefix match"           test_greedy_prefix_match
run_test "greedy substring match"        test_greedy_substring_match
run_test "greedy no match errors"        test_greedy_no_match_errors
run_test "greedy exact beats prefix"     test_greedy_exact_beats_prefix
run_test "delete nonexistent errors"     test_delete_nonexistent_errors
run_test "delete known session"          test_delete_known_session
run_test "delete alias rm"               test_delete_alias_rm
run_test "list shows session"            test_list_shows_session
run_test "list shows multiple"           test_list_shows_multiple
run_test "worktree create and path"      test_worktree_create_and_path

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
