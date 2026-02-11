#!/usr/bin/env bats
# Tests for core session management functionality

load setup

@test "sesh version displays version number" {
    run "$SESH_SCRIPT" version
    [ "$status" -eq 0 ]
    [[ "$output" =~ "sesh v" ]]
}

@test "sesh help displays usage information" {
    run "$SESH_SCRIPT" help
    [ "$status" -eq 0 ]
    [[ "$output" =~ "Usage:" ]]
    [[ "$output" =~ "Git Integration:" ]]
}

@test "sesh list shows no sessions initially" {
    run "$SESH_SCRIPT" list
    [ "$status" -eq 0 ]
    [[ "$output" =~ "Sessions" ]]
}

@test "session name validation rejects invalid names" {
    # Test with spaces
    run "$SESH_SCRIPT" new "invalid name"
    [ "$status" -eq 1 ]
    [[ "$output" =~ "must contain only" ]]
}

@test "session name validation rejects empty names" {
    run "$SESH_SCRIPT" new ""
    [ "$status" -eq 1 ]
    [[ "$output" =~ "cannot be empty" ]]
}

@test "session path is returned for existing session" {
    # Create a session directory manually for testing
    mkdir -p "$TEST_SESH_ROOT/test-session"
    
    run "$SESH_SCRIPT" test-session
    [ "$status" -eq 0 ]
    [[ "$output" == "$TEST_SESH_ROOT/test-session" ]]
}

@test "error is shown for non-existent session" {
    run "$SESH_SCRIPT" nonexistent
    [ "$status" -eq 1 ]
    [[ "$output" =~ "does not exist" ]]
}

@test "session rename works correctly" {
    # Create a session
    mkdir -p "$TEST_SESH_ROOT/old-name"
    echo "Test description" > "$TEST_SESH_ROOT/old-name/.sesh-desc"
    
    run "$SESH_SCRIPT" rename old-name new-name
    [ "$status" -eq 0 ]
    [ -d "$TEST_SESH_ROOT/new-name" ]
    [ ! -d "$TEST_SESH_ROOT/old-name" ]
    [ -f "$TEST_SESH_ROOT/new-name/.sesh-desc" ]
}

@test "describe command shows description" {
    mkdir -p "$TEST_SESH_ROOT/test-session"
    echo "Test description" > "$TEST_SESH_ROOT/test-session/.sesh-desc"
    
    run "$SESH_SCRIPT" describe test-session
    [ "$status" -eq 0 ]
    [[ "$output" == "Test description" ]]
}

@test "describe command updates description" {
    mkdir -p "$TEST_SESH_ROOT/test-session"
    
    run "$SESH_SCRIPT" describe test-session "New description"
    [ "$status" -eq 0 ]
    
    # Verify description was updated
    [ "$(cat "$TEST_SESH_ROOT/test-session/.sesh-desc")" = "New description" ]
}

@test "init zsh outputs shell integration code" {
    run "$SESH_SCRIPT" init zsh
    [ "$status" -eq 0 ]
    [[ "$output" =~ "sesh shell integration" ]]
    [[ "$output" =~ "function s()" ]]
    [[ "$output" =~ "function si()" ]]
    [[ "$output" =~ "compdef" ]]
}

@test "init rejects unsupported shells" {
    run "$SESH_SCRIPT" init bash
    [ "$status" -eq 1 ]
    [[ "$output" =~ "Only zsh is currently supported" ]]
}
