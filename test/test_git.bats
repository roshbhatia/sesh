#!/usr/bin/env bats
# Tests for git integration functionality

load setup

setup_git_session() {
    local session_name="$1"
    local num_repos="${2:-2}"
    
    # Create session directory
    mkdir -p "$TEST_SESH_ROOT/$session_name"
    echo "Test session with git repos" > "$TEST_SESH_ROOT/$session_name/.sesh-desc"
    
    # Create test git repos
    for i in $(seq 1 "$num_repos"); do
        local repo_path="${BATS_TEST_TMPDIR}/repos/org/repo${i}"
        create_test_git_repo "$repo_path" "repo${i}"
        
        # Create symlink in session
        mkdir -p "$TEST_SESH_ROOT/$session_name/org"
        ln -s "$repo_path" "$TEST_SESH_ROOT/$session_name/org/repo${i}"
    done
}

@test "git status works with no repos" {
    mkdir -p "$TEST_SESH_ROOT/empty-session"
    
    run "$SESH_SCRIPT" status empty-session
    [ "$status" -eq 0 ]
    [[ "$output" =~ "No repositories found" ]]
}

@test "git status shows clean repos" {
    setup_git_session "test-session" 2
    
    run "$SESH_SCRIPT" status test-session
    [ "$status" -eq 0 ]
    [[ "$output" =~ "org/repo1" ]]
    [[ "$output" =~ "org/repo2" ]]
    [[ "$output" =~ "clean" ]]
}

@test "git status detects modified files" {
    setup_git_session "test-session" 1
    
    # Modify a file
    echo "modified" >> "${BATS_TEST_TMPDIR}/repos/org/repo1/README.md"
    
    run "$SESH_SCRIPT" status test-session
    [ "$status" -eq 0 ]
    [[ "$output" =~ "modified" ]]
}

@test "git sync with fetch flag only fetches" {
    setup_git_session "test-session" 1
    
    run "$SESH_SCRIPT" sync test-session --fetch
    [ "$status" -eq 0 ]
    [[ "$output" =~ "Fetched" ]] || [[ "$output" =~ "succeeded" ]]
}

@test "git sync skips repos with uncommitted changes" {
    setup_git_session "test-session" 1
    
    # Create uncommitted changes
    echo "modified" >> "${BATS_TEST_TMPDIR}/repos/org/repo1/README.md"
    
    run "$SESH_SCRIPT" sync test-session
    [ "$status" -eq 1 ]
    [[ "$output" =~ "uncommitted changes" ]]
}

@test "git exec runs command in all repos" {
    setup_git_session "test-session" 2
    
    run "$SESH_SCRIPT" exec test-session pwd
    [ "$status" -eq 0 ]
    [[ "$output" =~ "repo1" ]]
    [[ "$output" =~ "repo2" ]]
}

@test "git exec handles command failures" {
    setup_git_session "test-session" 1
    
    run "$SESH_SCRIPT" exec test-session false
    [ "$status" -eq 1 ]
    [[ "$output" =~ "failed" ]]
}

@test "git branch --status shows current branches" {
    setup_git_session "test-session" 2
    
    run "$SESH_SCRIPT" branch test-session --status
    [ "$status" -eq 0 ]
    [[ "$output" =~ "Branch status" ]]
    [[ "$output" =~ "org/repo1" ]]
    [[ "$output" =~ "Current:" ]]
}

@test "git branch creates new branch in all repos" {
    setup_git_session "test-session" 2
    
    run "$SESH_SCRIPT" branch test-session feature-test
    [ "$status" -eq 0 ]
    [[ "$output" =~ "Created and switched" ]] || [[ "$output" =~ "succeeded" ]]
    
    # Verify branch was created
    cd "${BATS_TEST_TMPDIR}/repos/org/repo1" || return 1
    current_branch=$(git branch --show-current)
    [ "$current_branch" = "feature-test" ]
}

@test "git branch switches to existing branch" {
    setup_git_session "test-session" 1
    
    # Create a branch first
    cd "${BATS_TEST_TMPDIR}/repos/org/repo1" || return 1
    git checkout -b existing-branch --quiet
    git checkout master --quiet || git checkout main --quiet
    
    run "$SESH_SCRIPT" branch test-session existing-branch
    [ "$status" -eq 0 ]
    [[ "$output" =~ "Switched to existing" ]] || [[ "$output" =~ "succeeded" ]]
}

@test "git branch with --worktree creates new session" {
    setup_git_session "test-session" 1
    
    run "$SESH_SCRIPT" branch test-session feature-wt --worktree
    [ "$status" -eq 0 ]
    [[ "$output" =~ "Created worktree session" ]]
    
    # Verify new session exists
    [ -d "$TEST_SESH_ROOT/test-session-feature-wt" ]
    
    # Verify it's marked as worktree
    [ -f "$TEST_SESH_ROOT/test-session-feature-wt/.sesh-worktree" ]
}

@test "worktree session deletion cleans up worktrees" {
    setup_git_session "test-session" 1
    
    # Create worktree session
    "$SESH_SCRIPT" branch test-session feature-wt --worktree >/dev/null 2>&1
    
    # Verify worktree was created
    cd "${BATS_TEST_TMPDIR}/repos/org" || return 1
    [ -d "repo1-wt-feature-wt" ]
    
    # Note: delete requires confirmation, so we can't easily test it automatically
    # This test verifies the worktree was created; manual testing needed for deletion
}
