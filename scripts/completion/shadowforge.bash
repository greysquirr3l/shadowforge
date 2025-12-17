# shadowforge(1) completion                              -*- shell-script -*-

# Bash completion for shadowforge CLI
# Source this file or copy to /etc/bash_completion.d/

_shadowforge_completion() {
    local cur prev words cword
    _init_completion || return

    # Main commands
    local commands="embed extract analyze formats version help"

    # Global flags
    local global_flags="--help --verbose --debug --config"

    # Embed flags
    local embed_flags="-i --input -c --cover -o --output -t --technique -p --password -r --redundancy -q --quality -j --json"

    # Extract flags
    local extract_flags="-i --input -o --output -p --password -j --json"

    # Analyze flags
    local analyze_flags="-c --cover -t --technique -j --json"

    # Supported techniques
    local techniques="lsb dct phase echo zero-width lsb-audio palette"

    case "${words[1]}" in
        embed)
            case "${prev}" in
                -i|--input|-c|--cover|-o|--output)
                    # File completion
                    _filedir
                    return
                    ;;
                -t|--technique)
                    # Technique completion
                    COMPREPLY=( $(compgen -W "${techniques}" -- "${cur}") )
                    return
                    ;;
                -p|--password)
                    # No completion for password
                    return
                    ;;
                -r|--redundancy)
                    # Suggest common redundancy values
                    COMPREPLY=( $(compgen -W "0.2 0.3 0.5" -- "${cur}") )
                    return
                    ;;
                -q|--quality)
                    # Suggest common quality values
                    COMPREPLY=( $(compgen -W "80 90 95 100" -- "${cur}") )
                    return
                    ;;
                *)
                    COMPREPLY=( $(compgen -W "${embed_flags} ${global_flags}" -- "${cur}") )
                    return
                    ;;
            esac
            ;;
        extract)
            case "${prev}" in
                -i|--input|-o|--output)
                    _filedir
                    return
                    ;;
                -p|--password)
                    return
                    ;;
                *)
                    COMPREPLY=( $(compgen -W "${extract_flags} ${global_flags}" -- "${cur}") )
                    return
                    ;;
            esac
            ;;
        analyze)
            case "${words[2]}" in
                capacity)
                    case "${prev}" in
                        -c|--cover)
                            _filedir
                            return
                            ;;
                        -t|--technique)
                            COMPREPLY=( $(compgen -W "${techniques}" -- "${cur}") )
                            return
                            ;;
                        *)
                            COMPREPLY=( $(compgen -W "${analyze_flags} ${global_flags}" -- "${cur}") )
                            return
                            ;;
                    esac
                    ;;
                *)
                    COMPREPLY=( $(compgen -W "capacity" -- "${cur}") )
                    return
                    ;;
            esac
            ;;
        formats|version|help)
            COMPREPLY=( $(compgen -W "${global_flags}" -- "${cur}") )
            return
            ;;
        *)
            # If no command yet, suggest commands
            if [[ ${cword} -eq 1 ]]; then
                COMPREPLY=( $(compgen -W "${commands}" -- "${cur}") )
                return
            fi
            ;;
    esac
}

# Register completion
complete -F _shadowforge_completion shadowforge
complete -F _shadowforge_completion sforge

# vim: ft=bash sw=4 ts=4 et
