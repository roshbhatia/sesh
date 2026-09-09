export def --env sy [
  --greedy: string # Fuzzy-match a session name and print its path
  ...args: string@"__sy_completion_values_0"
] {
  if $greedy != null {
    return (^sy --greedy $greedy ...$args)
  }
  if (($args | length) == 1) and (not ($args.0 | str starts-with "-")) and ($args.0 not-in ["__values" "add" "archive" "at" "attach" "completion" "config" "cur" "current" "delete" "generate" "help" "info" "init" "list" "ls" "new" "open" "path" "provider" "remove" "rename" "restore" "rm" "source" "status" "sw" "switch" "unarchive"]) {
    let resolved = (^sy --greedy $args.0 | complete)
    if ($resolved.exit_code == 0) and (not ($resolved.stdout | str trim | is-empty)) {
      cd ($resolved.stdout | str trim)
      return
    }
  }
  ^sy ...$args
}

export extern "sy completion" [
  shell: string@"nu-complete sy shell"
]

def "nu-complete sy shell" [] { [bash zsh fish nu] }

export extern "sy add" [
  --branch(-b): string # Override branch name for all worktrees
  --stdin # Read repo paths from stdin
  ...args: string@"__sy_completion_values_1"
]

export extern "sy archive" [
  ...args: string@"__sy_completion_values_2"
]

export extern "sy attach" [
  ...args: string@"__sy_completion_values_3"
]

export extern "sy config" [
  ...args: string@"__sy_completion_none"
]

export extern "sy config edit" [
  ...args: string@"__sy_completion_none"
]

export extern "sy config init" [
  ...args: string@"__sy_completion_none"
]

export extern "sy current" [
  --path # Print the session path instead of its name
  --quiet(-q) # Print nothing when outside a session
  ...args: string@"__sy_completion_none"
]

export extern "sy delete" [
  --archived # Delete an archived session instead of an active one
  --force(-f) # Skip confirmation and delete even if worktree cleanup fails
  ...args: string@"__sy_completion_values_4"
]

export extern "sy help" [
  ...args: string@"__sy_completion_none"
]

export extern "sy init" [
  ...args: string@"__sy_completion_values_5"
]

export extern "sy list" [
  --archived # List archived sessions instead of active ones
  --format: string # Output format: table, json, names, paths
  --json # Output JSON
  --names # Output session names only
  --paths # Output session paths only
  ...args: string@"__sy_completion_none"
]

export extern "sy new" [
  --branch(-b): string # Override branch name for all worktrees
  --empty # Create the session with no repositories
  --stdin # Read repo paths from stdin
  ...args: string@"__sy_completion_values_6"
]

export extern "sy open" [
  --format: string # Output format: table, json
  ...args: string@"__sy_completion_values_7"
]

export extern "sy path" [
  ...args: string@"__sy_completion_values_8"
]

export extern "sy provider" [
  ...args: string@"__sy_completion_none"
]

export extern "sy remove" [
  --force(-f) # Skip confirmation prompt
  ...args: string@"__sy_completion_values_9"
]

export extern "sy rename" [
  ...args: string@"__sy_completion_values_10"
]

export extern "sy source" [
  ...args: string@"__sy_completion_none"
]

export extern "sy source list" [
  ...args: string@"__sy_completion_none"
]

export extern "sy status" [
  --format: string # Output format: table, json
  ...args: string@"__sy_completion_values_11"
]

export extern "sy switch" [
  --name # Print the resolved name instead of the path
  ...args: string@"__sy_completion_values_12"
]

export extern "sy unarchive" [
  ...args: string@"__sy_completion_values_13"
]

def "__sy_completion_none" [] { [] }

def "__sy_completion_values_0" [context?: string] {
  [
    "completion"
    "add"
    "archive"
    "attach"
    "config"
    "current"
    "delete"
    "help"
    "init"
    "list"
    "new"
    "open"
    "path"
    "provider"
    "remove"
    "rename"
    "source"
    "status"
    "switch"
    "unarchive"
    (try { run-external "sy" "__values" "active" ($context | default "") | lines } catch { [] })
  ] | flatten | uniq
}

def "__sy_completion_values_1" [context?: string] {
  [
    (try { run-external "sy" "__values" "add" ($context | default "") | lines } catch { [] })
  ] | flatten | uniq
}

def "__sy_completion_values_2" [context?: string] {
  [
    (try { run-external "sy" "__values" "active" ($context | default "") | lines } catch { [] })
  ] | flatten | uniq
}

def "__sy_completion_values_3" [context?: string] {
  [
    (try { run-external "sy" "__values" "active" ($context | default "") | lines } catch { [] })
  ] | flatten | uniq
}

def "__sy_completion_values_4" [context?: string] {
  [
    (try { run-external "sy" "__values" "sessions" ($context | default "") | lines } catch { [] })
  ] | flatten | uniq
}

def "__sy_completion_values_5" [context?: string] {
  [
    (try { run-external "sy" "__values" "shells" ($context | default "") | lines } catch { [] })
  ] | flatten | uniq
}

def "__sy_completion_values_6" [context?: string] {
  [
    (try { run-external "sy" "__values" "new" ($context | default "") | lines } catch { [] })
  ] | flatten | uniq
}

def "__sy_completion_values_7" [context?: string] {
  [
    (try { run-external "sy" "__values" "active" ($context | default "") | lines } catch { [] })
  ] | flatten | uniq
}

def "__sy_completion_values_8" [context?: string] {
  [
    (try { run-external "sy" "__values" "active" ($context | default "") | lines } catch { [] })
  ] | flatten | uniq
}

def "__sy_completion_values_9" [context?: string] {
  [
    (try { run-external "sy" "__values" "remove" ($context | default "") | lines } catch { [] })
  ] | flatten | uniq
}

def "__sy_completion_values_10" [context?: string] {
  [
    (try { run-external "sy" "__values" "active" ($context | default "") | lines } catch { [] })
  ] | flatten | uniq
}

def "__sy_completion_values_11" [context?: string] {
  [
    (try { run-external "sy" "__values" "active" ($context | default "") | lines } catch { [] })
  ] | flatten | uniq
}

def "__sy_completion_values_12" [context?: string] {
  [
    (try { run-external "sy" "__values" "active" ($context | default "") | lines } catch { [] })
  ] | flatten | uniq
}

def "__sy_completion_values_13" [context?: string] {
  [
    (try { run-external "sy" "__values" "archived" ($context | default "") | lines } catch { [] })
  ] | flatten | uniq
}
