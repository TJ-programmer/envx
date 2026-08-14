package cli

import (
	"errors"
	"fmt"
	"io"
)

const bashCompletion = `_envx_completions() {
    local cur="${COMP_WORDS[COMP_CWORD]}"
    local commands="init set get list unset run env import export diff doctor config key web completions help"
    if [ "$COMP_CWORD" -eq 1 ]; then
        COMPREPLY=( $(compgen -W "$commands" -- "$cur") )
        return 0
    fi
    case "${COMP_WORDS[1]}" in
        env)    COMPREPLY=( $(compgen -W "create use delete list" -- "$cur") ) ;;
        config) COMPREPLY=( $(compgen -W "get set" -- "$cur") ) ;;
        key)    COMPREPLY=( $(compgen -W "status rotate export import" -- "$cur") ) ;;
        web)    COMPREPLY=( $(compgen -W "--port --no-open" -- "$cur") ) ;;
        *)      COMPREPLY=( $(compgen -W "--env --format --secret --plain --show-secret --force --no-gitignore --shell --overlay --watch --root --help" -- "$cur") ) ;;
    esac
}
complete -F _envx_completions envx
`

const zshCompletion = `#compdef envx
_arguments '1:command:(init set get list unset run env import export diff doctor config key web completions help)' '*::arg:->args'
`

const fishCompletion = `complete -c envx -f -a "init set get list unset run env import export diff doctor config key web completions help"
complete -c envx -f -n "__fish_use_subcommand" -a "env" -d "Manage environments"
`

const powershellCompletion = `Register-ArgumentCompleter -Native -CommandName envx -ScriptBlock {
    param($wordToComplete, $commandAst, $cursorPosition)
    'init','set','get','list','unset','run','env','import','export','diff','doctor','config','key','web','completions','help' |
        Where-Object { $_ -like "$wordToComplete*" } |
        ForEach-Object {
            [System.Management.Automation.CompletionResult]::new($_, $_, 'ParameterValue', $_)
        }
}
`

func cmdCompletions(args []string, stdout, stderr io.Writer) int {
	if checkHelp(args, stdout, "completions") {
		return 0
	}
	if len(args) < 1 {
		return printError(stderr, errors.New("usage: envx completions bash|zsh|fish|powershell"))
	}
	switch args[0] {
	case "bash":
		fmt.Fprint(stdout, bashCompletion)
	case "zsh":
		fmt.Fprint(stdout, zshCompletion)
	case "fish":
		fmt.Fprint(stdout, fishCompletion)
	case "powershell":
		fmt.Fprint(stdout, powershellCompletion)
	default:
		return printError(stderr, fmt.Errorf("unknown shell '%s' (use bash, zsh, fish, or powershell)", args[0]))
	}
	return 0
}
