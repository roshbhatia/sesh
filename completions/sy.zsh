#compdef sy
__sy_completion_values_0() {
  local -a values
  values=( 'completion' 'add' 'archive' 'attach' 'config' 'current' 'delete' 'help' 'init' 'list' 'new' 'open' 'path' 'provider' 'prune' 'remove' 'rename' 'source' 'status' 'switch' 'unarchive')
  values+=("${(@f)$('sy' '__values' 'active' "${BUFFER[1,CURSOR]}" 2>/dev/null)}")
  compadd -a values
}
__sy_completion_values_1() {
  local -a values
  values=()
  values+=("${(@f)$('sy' '__values' 'add' "${BUFFER[1,CURSOR]}" 2>/dev/null)}")
  compadd -a values
}
__sy_completion_values_2() {
  local -a values
  values=()
  values+=("${(@f)$('sy' '__values' 'active' "${BUFFER[1,CURSOR]}" 2>/dev/null)}")
  compadd -a values
}
__sy_completion_values_3() {
  local -a values
  values=()
  values+=("${(@f)$('sy' '__values' 'active' "${BUFFER[1,CURSOR]}" 2>/dev/null)}")
  compadd -a values
}
__sy_completion_values_4() {
  local -a values
  values=()
  values+=("${(@f)$('sy' '__values' 'sessions' "${BUFFER[1,CURSOR]}" 2>/dev/null)}")
  compadd -a values
}
__sy_completion_values_5() {
  local -a values
  values=()
  values+=("${(@f)$('sy' '__values' 'shells' "${BUFFER[1,CURSOR]}" 2>/dev/null)}")
  compadd -a values
}
__sy_completion_values_6() {
  local -a values
  values=()
  values+=("${(@f)$('sy' '__values' 'new' "${BUFFER[1,CURSOR]}" 2>/dev/null)}")
  compadd -a values
}
__sy_completion_values_7() {
  local -a values
  values=()
  values+=("${(@f)$('sy' '__values' 'active' "${BUFFER[1,CURSOR]}" 2>/dev/null)}")
  compadd -a values
}
__sy_completion_values_8() {
  local -a values
  values=()
  values+=("${(@f)$('sy' '__values' 'active' "${BUFFER[1,CURSOR]}" 2>/dev/null)}")
  compadd -a values
}
__sy_completion_values_9() {
  local -a values
  values=()
  values+=("${(@f)$('sy' '__values' 'remove' "${BUFFER[1,CURSOR]}" 2>/dev/null)}")
  compadd -a values
}
__sy_completion_values_10() {
  local -a values
  values=()
  values+=("${(@f)$('sy' '__values' 'active' "${BUFFER[1,CURSOR]}" 2>/dev/null)}")
  compadd -a values
}
__sy_completion_values_11() {
  local -a values
  values=()
  values+=("${(@f)$('sy' '__values' 'active' "${BUFFER[1,CURSOR]}" 2>/dev/null)}")
  compadd -a values
}
__sy_completion_values_12() {
  local -a values
  values=()
  values+=("${(@f)$('sy' '__values' 'active' "${BUFFER[1,CURSOR]}" 2>/dev/null)}")
  compadd -a values
}
__sy_completion_values_13() {
  local -a values
  values=()
  values+=("${(@f)$('sy' '__values' 'archived' "${BUFFER[1,CURSOR]}" 2>/dev/null)}")
  compadd -a values
}

_sy() {
  local context=''
  local word
  local consume_value=0
  local options_done=0
  for word in ${words[2,$((CURRENT - 1))]}; do
    if (( consume_value )); then
      consume_value=0
      continue
    fi
    if (( options_done )); then
      continue
    fi
    if [[ "$word" == '--' ]]; then
      options_done=1
      continue
    fi
    case "$context:$word" in
      ':--config') consume_value=1; continue ;;
      ':--config='*) continue ;;
      ':--greedy') consume_value=1; continue ;;
      ':--greedy='*) continue ;;
      'add:--branch') consume_value=1; continue ;;
      'add:--branch='*) continue ;;
      'add:-b') consume_value=1; continue ;;
      'add:-b='*) continue ;;
      'add:--start-point') consume_value=1; continue ;;
      'add:--start-point='*) continue ;;
      'list:--format') consume_value=1; continue ;;
      'list:--format='*) continue ;;
      'new:--branch') consume_value=1; continue ;;
      'new:--branch='*) continue ;;
      'new:-b') consume_value=1; continue ;;
      'new:-b='*) continue ;;
      'new:--start-point') consume_value=1; continue ;;
      'new:--start-point='*) continue ;;
      'open:--format') consume_value=1; continue ;;
      'open:--format='*) continue ;;
      'status:--format') consume_value=1; continue ;;
      'status:--format='*) continue ;;
    esac
    case "$context:$word" in
      ':completion') context='completion' ;;
      ':add') context='add' ;;
      ':archive') context='archive' ;;
      ':attach') context='attach' ;;
      ':config') context='config' ;;
      'config:edit') context='config edit' ;;
      'config:init') context='config init' ;;
      ':current') context='current' ;;
      ':delete') context='delete' ;;
      ':help') context='help' ;;
      ':init') context='init' ;;
      ':list') context='list' ;;
      ':new') context='new' ;;
      ':open') context='open' ;;
      ':path') context='path' ;;
      ':provider') context='provider' ;;
      ':prune') context='prune' ;;
      ':remove') context='remove' ;;
      ':rename') context='rename' ;;
      ':source') context='source' ;;
      'source:list') context='source list' ;;
      ':status') context='status' ;;
      ':switch') context='switch' ;;
      ':unarchive') context='unarchive' ;;
    esac
  done
  case "$context" in
    '')
      _arguments \
        '--config[Config file, instead of $SESHY_CONFIG or $XDG_CONFIG_HOME/seshy/config.yaml]:value:' \
        '--greedy[Fuzzy-match a session name and print its path]:value:' \
        '*:argument:__sy_completion_values_0'

      ;;
    'completion')
      _arguments \
        '2:shell:(bash zsh fish nu)'
      ;;
    'add')
      _arguments \
        '(-b)--branch[Override branch name for all worktrees]:value:' \
        '--existing[Check out the existing branch named by --branch]' \
        '--reference[Link existing directories without creating worktrees]' \
        '--start-point[Commit to start a new branch from (default HEAD)]:value:' \
        '--stdin[Read repo paths from stdin]' \
        '*:argument:__sy_completion_values_1'

      ;;
    'archive')
      _arguments \
        '*:argument:__sy_completion_values_2'

      ;;
    'attach')
      _arguments \
        '*:argument:__sy_completion_values_3'

      ;;
    'config')
      _arguments \
        '2:command:(edit init)'

      ;;
    'config edit')
      _arguments \
        '*:argument:'

      ;;
    'config init')
      _arguments \
        '*:argument:'

      ;;
    'current')
      _arguments \
        '--path[Print the session path instead of its name]' \
        '(-q)--quiet[Print nothing when outside a session]' \
        '*:argument:'

      ;;
    'delete')
      _arguments \
        '--archived[Delete an archived session instead of an active one]' \
        '(-f)--force[Skip confirmation and delete even if worktree cleanup fails]' \
        '(-y)--yes[Skip the confirmation prompt]' \
        '*:argument:__sy_completion_values_4'

      ;;
    'help')
      _arguments \
        '*:argument:'

      ;;
    'init')
      _arguments \
        '*:argument:__sy_completion_values_5'

      ;;
    'list')
      _arguments \
        '--archived[List archived sessions instead of active ones]' \
        '--format[Output format: table, json, names, paths]:value:' \
        '--json[Output JSON]' \
        '--names[Output session names only]' \
        '--paths[Output session paths only]' \
        '*:argument:'

      ;;
    'new')
      _arguments \
        '(-b)--branch[Override branch name for all worktrees]:value:' \
        '--empty[Create the session with no repositories]' \
        '--existing[Check out the existing branch named by --branch]' \
        '--reference[Link existing directories without creating worktrees]' \
        '--start-point[Commit to start a new branch from (default HEAD)]:value:' \
        '--stdin[Read repo paths from stdin]' \
        '*:argument:__sy_completion_values_6'

      ;;
    'open')
      _arguments \
        '--format[Output format: table, json]:value:' \
        '*:argument:__sy_completion_values_7'

      ;;
    'path')
      _arguments \
        '*:argument:__sy_completion_values_8'

      ;;
    'provider')
      _arguments \
        '*:argument:'

      ;;
    'prune')
      _arguments \
        '--dry-run[Print the actions without taking them]' \
        '*:argument:'

      ;;
    'remove')
      _arguments \
        '(-f)--force[Skip confirmation prompt and remove even if worktree cleanup fails]' \
        '(-y)--yes[Skip the confirmation prompt]' \
        '*:argument:__sy_completion_values_9'

      ;;
    'rename')
      _arguments \
        '*:argument:__sy_completion_values_10'

      ;;
    'source')
      _arguments \
        '2:command:(list)'

      ;;
    'source list')
      _arguments \
        '*:argument:'

      ;;
    'status')
      _arguments \
        '--format[Output format: table, json]:value:' \
        '*:argument:__sy_completion_values_11'

      ;;
    'switch')
      _arguments \
        '--name[Print the resolved name instead of the path]' \
        '*:argument:__sy_completion_values_12'

      ;;
    'unarchive')
      _arguments \
        '*:argument:__sy_completion_values_13'

      ;;
  esac
}
compdef _sy sy
