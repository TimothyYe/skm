_get_skm_dirs() {
    ls ~/.skm
}

_get_skm_commands() {
    echo "init create ls use delete rename copy display backup restore cache help"
}

_skm_autocomplete() {
    local cur prev opts

    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"

    case "${prev}" in
        init|i)
            COMPREPLY=( $(compgen -W "" -- ${cur}) )
            ;;
        create|c)
            COMPREPLY=( $(compgen -W "" -- ${cur}) )
            ;;
        ls|l)
            COMPREPLY=( $(compgen -W "" -- ${cur}) )
            ;;
        use|u)
            COMPREPLY=( $(compgen -W "$(_get_skm_dirs)" -- ${cur}) )
            ;;
        delete|d)
            COMPREPLY=( $(compgen -W "$(_get_skm_dirs)" -- ${cur}) )
            ;;
        rename|rn)
            COMPREPLY=( $(compgen -W "$(_get_skm_dirs)" -- ${cur}) )
            ;;
        copy|cp)
            COMPREPLY=( $(compgen -W "--add --del --list" -- ${cur}) )
            ;;
        display|dp)
            COMPREPLY=( $(compgen -W "$(_get_skm_dirs)" -- ${cur}) )
            ;;
        backup|b)
            COMPREPLY=( $(compgen -W "--restic" -- ${cur}) )
            ;;
        restore|r)
            COMPREPLY=( $(compgen -W "--restic  --restic-snapshot" -- ${cur}) )
            ;;
        cache)
            COMPREPLY=( $(compgen -W "$(_get_skm_dirs)" -- ${cur}) )
            ;;
        help|h)
            COMPREPLY=( $(compgen -W "$(_get_skm_commands)" -- ${cur}) )
            ;;
        *)
            COMPREPLY=( $(compgen -W "$(_get_skm_commands)" -- ${cur}) )
            ;;
    esac

    return 0
}

complete -F _skm_autocomplete skm

