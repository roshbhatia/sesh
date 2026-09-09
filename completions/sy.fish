complete -c sy -e
complete -c sy -f
function __sy_completion_values_0
  begin
    printf '%s\n' 'completion' 'add' 'archive' 'attach' 'config' 'current' 'delete' 'help' 'init' 'list' 'new' 'open' 'path' 'remove' 'rename' 'status' 'switch' 'unarchive'
    command 'sy' '__values' 'active' (commandline -cp) 2>/dev/null; or true
  end | string match -rv '\t'; or true
end
function __sy_completion_values_1
  begin
    command 'sy' '__values' 'add' (commandline -cp) 2>/dev/null; or true
  end | string match -rv '\t'; or true
end
function __sy_completion_values_2
  begin
    command 'sy' '__values' 'active' (commandline -cp) 2>/dev/null; or true
  end | string match -rv '\t'; or true
end
function __sy_completion_values_3
  begin
    command 'sy' '__values' 'active' (commandline -cp) 2>/dev/null; or true
  end | string match -rv '\t'; or true
end
function __sy_completion_values_4
  begin
    command 'sy' '__values' 'sessions' (commandline -cp) 2>/dev/null; or true
  end | string match -rv '\t'; or true
end
function __sy_completion_values_5
  begin
    command 'sy' '__values' 'shells' (commandline -cp) 2>/dev/null; or true
  end | string match -rv '\t'; or true
end
function __sy_completion_values_6
  begin
    command 'sy' '__values' 'new' (commandline -cp) 2>/dev/null; or true
  end | string match -rv '\t'; or true
end
function __sy_completion_values_7
  begin
    command 'sy' '__values' 'active' (commandline -cp) 2>/dev/null; or true
  end | string match -rv '\t'; or true
end
function __sy_completion_values_8
  begin
    command 'sy' '__values' 'active' (commandline -cp) 2>/dev/null; or true
  end | string match -rv '\t'; or true
end
function __sy_completion_values_9
  begin
    command 'sy' '__values' 'remove' (commandline -cp) 2>/dev/null; or true
  end | string match -rv '\t'; or true
end
function __sy_completion_values_10
  begin
    command 'sy' '__values' 'active' (commandline -cp) 2>/dev/null; or true
  end | string match -rv '\t'; or true
end
function __sy_completion_values_11
  begin
    command 'sy' '__values' 'active' (commandline -cp) 2>/dev/null; or true
  end | string match -rv '\t'; or true
end
function __sy_completion_values_12
  begin
    command 'sy' '__values' 'active' (commandline -cp) 2>/dev/null; or true
  end | string match -rv '\t'; or true
end
function __sy_completion_values_13
  begin
    command 'sy' '__values' 'archived' (commandline -cp) 2>/dev/null; or true
  end | string match -rv '\t'; or true
end

function __sy_completion_context
  set -l context ''
  set -l words (commandline -opc)
  set -l consume_value 0
  set -l options_done 0
  for word in $words[2..-1]
    if test $consume_value -eq 1
      set consume_value 0
      continue
    end
    if test $options_done -eq 1
      continue
    end
    if test "$word" = '--'
      set options_done 1
      continue
    end
    switch "$context:$word"
      case ':--greedy'
        set consume_value 1
        continue
      case ':--greedy=*'
        continue
      case 'add:--branch' 'add:-b'
        set consume_value 1
        continue
      case 'add:--branch=*'
        continue
      case 'add:-b=*'
        continue
      case 'list:--format'
        set consume_value 1
        continue
      case 'list:--format=*'
        continue
      case 'new:--branch' 'new:-b'
        set consume_value 1
        continue
      case 'new:--branch=*'
        continue
      case 'new:-b=*'
        continue
      case 'open:--format'
        set consume_value 1
        continue
      case 'open:--format=*'
        continue
    end
    switch "$context:$word"
      case ':completion'
        set context 'completion'
      case ':add'
        set context 'add'
      case ':archive'
        set context 'archive'
      case ':attach'
        set context 'attach'
      case ':config'
        set context 'config'
      case 'config:edit'
        set context 'config edit'
      case 'config:init'
        set context 'config init'
      case ':current'
        set context 'current'
      case ':delete'
        set context 'delete'
      case ':help'
        set context 'help'
      case ':init'
        set context 'init'
      case ':list'
        set context 'list'
      case ':new'
        set context 'new'
      case ':open'
        set context 'open'
      case ':path'
        set context 'path'
      case ':remove'
        set context 'remove'
      case ':rename'
        set context 'rename'
      case ':status'
        set context 'status'
      case ':switch'
        set context 'switch'
      case ':unarchive'
        set context 'unarchive'
    end
  end
  echo $context
end
complete -c sy -n 'test (__sy_completion_context) = ""' -l greedy -r -d 'Fuzzy-match a session name and print its path'
complete -c sy -f -n 'test (__sy_completion_context) = ""' -a completion -d 'Generate shell completions'
complete -c sy -f -n 'test (__sy_completion_context) = ""' -a add -d 'sy add <name> [repos...] [flags]'
complete -c sy -f -n 'test (__sy_completion_context) = ""' -a archive -d 'sy archive [name]'
complete -c sy -f -n 'test (__sy_completion_context) = ""' -a attach -d 'sy attach <name>'
complete -c sy -f -n 'test (__sy_completion_context) = ""' -a config -d 'sy config'
complete -c sy -f -n 'test (__sy_completion_context) = ""' -a current -d 'sy current [flags]'
complete -c sy -f -n 'test (__sy_completion_context) = ""' -a delete -d 'sy delete [name] [flags]'
complete -c sy -f -n 'test (__sy_completion_context) = ""' -a help -d 'sy help [command]'
complete -c sy -f -n 'test (__sy_completion_context) = ""' -a init -d 'sy init <shell>'
complete -c sy -f -n 'test (__sy_completion_context) = ""' -a list -d 'sy list [flags]'
complete -c sy -f -n 'test (__sy_completion_context) = ""' -a new -d 'sy new <name> [repos...] [flags]'
complete -c sy -f -n 'test (__sy_completion_context) = ""' -a open -d 'sy open <name> [flags]'
complete -c sy -f -n 'test (__sy_completion_context) = ""' -a path -d 'sy path <name>'
complete -c sy -f -n 'test (__sy_completion_context) = ""' -a remove -d 'sy remove <session> <repo> [flags]'
complete -c sy -f -n 'test (__sy_completion_context) = ""' -a rename -d 'sy rename <old-name> <new-name>'
complete -c sy -f -n 'test (__sy_completion_context) = ""' -a status -d 'sy status [name]'
complete -c sy -f -n 'test (__sy_completion_context) = ""' -a switch -d 'sy switch <name> [flags]'
complete -c sy -f -n 'test (__sy_completion_context) = ""' -a unarchive -d 'sy unarchive [name]'
complete -c sy -f -n 'test (__sy_completion_context) = ""' -a '(__sy_completion_values_0)'
complete -c sy -f -n 'test (__sy_completion_context) = "completion"' -a 'bash zsh fish nu'
complete -c sy -n 'test (__sy_completion_context) = "add"' -l branch -s b -r -d 'Override branch name for all worktrees'
complete -c sy -n 'test (__sy_completion_context) = "add"' -l stdin -d 'Read repo paths from stdin'
complete -c sy -f -n 'test (__sy_completion_context) = "add"' -a '(__sy_completion_values_1)'
complete -c sy -f -n 'test (__sy_completion_context) = "archive"' -a '(__sy_completion_values_2)'
complete -c sy -f -n 'test (__sy_completion_context) = "attach"' -a '(__sy_completion_values_3)'
complete -c sy -f -n 'test (__sy_completion_context) = "config"' -a edit -d 'sy config edit'
complete -c sy -f -n 'test (__sy_completion_context) = "config"' -a init -d 'sy config init'
complete -c sy -n 'test (__sy_completion_context) = "current"' -l path -d 'Print the session path instead of its name'
complete -c sy -n 'test (__sy_completion_context) = "current"' -l quiet -s q -d 'Print nothing when outside a session'
complete -c sy -n 'test (__sy_completion_context) = "delete"' -l archived -d 'Delete an archived session instead of an active one'
complete -c sy -n 'test (__sy_completion_context) = "delete"' -l force -s f -d 'Skip confirmation and delete even if worktree cleanup fails'
complete -c sy -f -n 'test (__sy_completion_context) = "delete"' -a '(__sy_completion_values_4)'
complete -c sy -f -n 'test (__sy_completion_context) = "init"' -a '(__sy_completion_values_5)'
complete -c sy -n 'test (__sy_completion_context) = "list"' -l archived -d 'List archived sessions instead of active ones'
complete -c sy -n 'test (__sy_completion_context) = "list"' -l format -r -d 'Output format: table, json, names, paths'
complete -c sy -n 'test (__sy_completion_context) = "list"' -l json -d 'Output JSON'
complete -c sy -n 'test (__sy_completion_context) = "list"' -l names -d 'Output session names only'
complete -c sy -n 'test (__sy_completion_context) = "list"' -l paths -d 'Output session paths only'
complete -c sy -n 'test (__sy_completion_context) = "new"' -l branch -s b -r -d 'Override branch name for all worktrees'
complete -c sy -n 'test (__sy_completion_context) = "new"' -l empty -d 'Create the session with no repositories'
complete -c sy -n 'test (__sy_completion_context) = "new"' -l stdin -d 'Read repo paths from stdin'
complete -c sy -f -n 'test (__sy_completion_context) = "new"' -a '(__sy_completion_values_6)'
complete -c sy -n 'test (__sy_completion_context) = "open"' -l format -r -d 'Output format: table, json'
complete -c sy -f -n 'test (__sy_completion_context) = "open"' -a '(__sy_completion_values_7)'
complete -c sy -f -n 'test (__sy_completion_context) = "path"' -a '(__sy_completion_values_8)'
complete -c sy -n 'test (__sy_completion_context) = "remove"' -l force -s f -d 'Skip confirmation prompt'
complete -c sy -f -n 'test (__sy_completion_context) = "remove"' -a '(__sy_completion_values_9)'
complete -c sy -f -n 'test (__sy_completion_context) = "rename"' -a '(__sy_completion_values_10)'
complete -c sy -f -n 'test (__sy_completion_context) = "status"' -a '(__sy_completion_values_11)'
complete -c sy -n 'test (__sy_completion_context) = "switch"' -l name -d 'Print the resolved name instead of the path'
complete -c sy -f -n 'test (__sy_completion_context) = "switch"' -a '(__sy_completion_values_12)'
complete -c sy -f -n 'test (__sy_completion_context) = "unarchive"' -a '(__sy_completion_values_13)'
