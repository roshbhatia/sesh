__sy_completion_values_0() {
  printf '%s\n' 'completion' 'add' 'archive' 'attach' 'config' 'current' 'delete' 'help' 'init' 'list' 'new' 'open' 'path' 'provider' 'prune' 'remove' 'rename' 'source' 'status' 'switch' 'unarchive'
  'sy' '__values' 'active' "${COMP_LINE:0:COMP_POINT}" 2>/dev/null || true
}
__sy_completion_values_1() {
  'sy' '__values' 'add' "${COMP_LINE:0:COMP_POINT}" 2>/dev/null || true
}
__sy_completion_values_2() {
  'sy' '__values' 'active' "${COMP_LINE:0:COMP_POINT}" 2>/dev/null || true
}
__sy_completion_values_3() {
  'sy' '__values' 'active' "${COMP_LINE:0:COMP_POINT}" 2>/dev/null || true
}
__sy_completion_values_4() {
  'sy' '__values' 'sessions' "${COMP_LINE:0:COMP_POINT}" 2>/dev/null || true
}
__sy_completion_values_5() {
  'sy' '__values' 'shells' "${COMP_LINE:0:COMP_POINT}" 2>/dev/null || true
}
__sy_completion_values_6() {
  'sy' '__values' 'new' "${COMP_LINE:0:COMP_POINT}" 2>/dev/null || true
}
__sy_completion_values_7() {
  'sy' '__values' 'active' "${COMP_LINE:0:COMP_POINT}" 2>/dev/null || true
}
__sy_completion_values_8() {
  'sy' '__values' 'active' "${COMP_LINE:0:COMP_POINT}" 2>/dev/null || true
}
__sy_completion_values_9() {
  'sy' '__values' 'remove' "${COMP_LINE:0:COMP_POINT}" 2>/dev/null || true
}
__sy_completion_values_10() {
  'sy' '__values' 'active' "${COMP_LINE:0:COMP_POINT}" 2>/dev/null || true
}
__sy_completion_values_11() {
  'sy' '__values' 'active' "${COMP_LINE:0:COMP_POINT}" 2>/dev/null || true
}
__sy_completion_values_12() {
  'sy' '__values' 'active' "${COMP_LINE:0:COMP_POINT}" 2>/dev/null || true
}
__sy_completion_values_13() {
  'sy' '__values' 'archived' "${COMP_LINE:0:COMP_POINT}" 2>/dev/null || true
}
__sy_completion_filter() {
  local prefix="$1"
  local prepend="${2-}"
  local candidate
  local existing
  local duplicate
  COMPREPLY=()
  while IFS= read -r candidate || [[ -n "$candidate" ]]; do
    [[ "$candidate" == "$prefix"* ]] || continue
    candidate="$prepend$candidate"
    duplicate=0
    for existing in "${COMPREPLY[@]}"; do
      if [[ "$existing" == "$candidate" ]]; then
        duplicate=1
        break
      fi
    done
    (( duplicate )) || COMPREPLY+=("$candidate")
  done
}

_sy_complete() {
  local current="${COMP_WORDS[COMP_CWORD]}"
  local previous=""
  local context=""
  local word
  local index
  local consume_value=0
  local options_done=0
  if (( COMP_CWORD > 0 )); then
    previous="${COMP_WORDS[COMP_CWORD-1]}"
  fi
  for ((index=1; index<COMP_CWORD; index++)); do
    word="${COMP_WORDS[index]}"
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
      ':--greedy') consume_value=1; continue ;;
      ':--greedy='*) continue ;;
      'add:--branch') consume_value=1; continue ;;
      'add:--branch='*) continue ;;
      'add:-b') consume_value=1; continue ;;
      'add:-b='*) continue ;;
      'list:--format') consume_value=1; continue ;;
      'list:--format='*) continue ;;
      'new:--branch') consume_value=1; continue ;;
      'new:--branch='*) continue ;;
      'new:-b') consume_value=1; continue ;;
      'new:-b='*) continue ;;
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
  case "$context:$previous" in
  esac
  case "$context:$current" in
  esac
  case "$context" in
    '')
      __sy_completion_filter "$current" < <(
        printf '%s\n' 'completion' 'add' 'archive' 'attach' 'config' 'current' 'delete' 'help' 'init' 'list' 'new' 'open' 'path' 'provider' 'prune' 'remove' 'rename' 'source' 'status' 'switch' 'unarchive' '--greedy'
        __sy_completion_values_0
      )
      ;;
    'completion')
      __sy_completion_filter "$current" < <(
        printf '%s\n' 'bash' 'zsh' 'fish' 'nu'
      )
      ;;
    'add')
      __sy_completion_filter "$current" < <(
        printf '%s\n' '--branch' '-b' '--stdin'
        __sy_completion_values_1
      )
      ;;
    'archive')
      __sy_completion_filter "$current" < <(
        __sy_completion_values_2
      )
      ;;
    'attach')
      __sy_completion_filter "$current" < <(
        __sy_completion_values_3
      )
      ;;
    'config')
      __sy_completion_filter "$current" < <(
        printf '%s\n' 'edit' 'init'
      )
      ;;
    'config edit')
      __sy_completion_filter "$current" < <(
      )
      ;;
    'config init')
      __sy_completion_filter "$current" < <(
      )
      ;;
    'current')
      __sy_completion_filter "$current" < <(
        printf '%s\n' '--path' '--quiet' '-q'
      )
      ;;
    'delete')
      __sy_completion_filter "$current" < <(
        printf '%s\n' '--archived' '--force' '-f'
        __sy_completion_values_4
      )
      ;;
    'help')
      __sy_completion_filter "$current" < <(
      )
      ;;
    'init')
      __sy_completion_filter "$current" < <(
        __sy_completion_values_5
      )
      ;;
    'list')
      __sy_completion_filter "$current" < <(
        printf '%s\n' '--archived' '--format' '--json' '--names' '--paths'
      )
      ;;
    'new')
      __sy_completion_filter "$current" < <(
        printf '%s\n' '--branch' '-b' '--empty' '--stdin'
        __sy_completion_values_6
      )
      ;;
    'open')
      __sy_completion_filter "$current" < <(
        printf '%s\n' '--format'
        __sy_completion_values_7
      )
      ;;
    'path')
      __sy_completion_filter "$current" < <(
        __sy_completion_values_8
      )
      ;;
    'provider')
      __sy_completion_filter "$current" < <(
      )
      ;;
    'prune')
      __sy_completion_filter "$current" < <(
        printf '%s\n' '--dry-run'
      )
      ;;
    'remove')
      __sy_completion_filter "$current" < <(
        printf '%s\n' '--force' '-f'
        __sy_completion_values_9
      )
      ;;
    'rename')
      __sy_completion_filter "$current" < <(
        __sy_completion_values_10
      )
      ;;
    'source')
      __sy_completion_filter "$current" < <(
        printf '%s\n' 'list'
      )
      ;;
    'source list')
      __sy_completion_filter "$current" < <(
      )
      ;;
    'status')
      __sy_completion_filter "$current" < <(
        printf '%s\n' '--format'
        __sy_completion_values_11
      )
      ;;
    'switch')
      __sy_completion_filter "$current" < <(
        printf '%s\n' '--name'
        __sy_completion_values_12
      )
      ;;
    'unarchive')
      __sy_completion_filter "$current" < <(
        __sy_completion_values_13
      )
      ;;
  esac
}
complete -F _sy_complete sy
