# Common test setup for bats tests

# Set up test environment
export TEST_SESH_ROOT="${BATS_TEST_TMPDIR}/sesh-test"
export SESH_ROOT="$TEST_SESH_ROOT"
export XDG_STATE_HOME="${BATS_TEST_TMPDIR}/state"

# Path to sesh script
SESH_SCRIPT="${BATS_TEST_DIRNAME}/../sesh"

# Create test git repo helper
create_test_git_repo() {
  local repo_path="$1"
  local repo_name="${2:-test-repo}"

  mkdir -p "$repo_path"
  cd "$repo_path" || return 1

  git init --quiet
  git config user.email "test@example.com"
  git config user.name "Test User"

  echo "# $repo_name" >README.md
  git add README.md
  git commit -m "Initial commit" --quiet

  cd - >/dev/null || return 1
}

# Setup function called before each test
setup() {
  # Create clean test directories
  mkdir -p "$TEST_SESH_ROOT"
  mkdir -p "$XDG_STATE_HOME"
  mkdir -p "${BATS_TEST_TMPDIR}/repos/org"
}

# Teardown function called after each test
teardown() {
  # Clean up test environment
  rm -rf "$TEST_SESH_ROOT"
  rm -rf "$XDG_STATE_HOME"
  rm -rf "${BATS_TEST_TMPDIR}/repos"
}
