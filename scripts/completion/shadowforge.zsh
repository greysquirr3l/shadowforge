#compdef shadowforge sforge

# Zsh completion for shadowforge CLI

_shadowforge() {
    local -a commands
    commands=(
        'embed:Embed payload into cover media'
        'extract:Extract payload from stego media'
        'analyze:Analyze media capacity'
        'formats:List supported techniques'
        'version:Show version information'
        'help:Show help information'
    )

    local -a global_flags
    global_flags=(
        '--help[Show help information]'
        '--verbose[Enable verbose logging]'
        '--debug[Enable debug logging]'
        '--config[Configuration file path]:file:_files'
    )

    local -a techniques
    techniques=(
        'lsb:Least Significant Bit (Image)'
        'dct:Discrete Cosine Transform (JPEG)'
        'phase:Phase Encoding (Audio)'
        'echo:Echo Hiding (Audio)'
        'zero-width:Zero-Width Characters (Text)'
        'lsb-audio:LSB Audio (WAV)'
        'palette:Palette-Based (GIF/PNG)'
    )

    _arguments -C \
        "1: :{_describe 'command' commands}" \
        '*:: :->args' \
        $global_flags

    case $state in
        args)
            case $words[1] in
                embed)
                    _arguments \
                        '(-i --input)'{-i,--input}'[Input file to embed]:file:_files' \
                        '(-c --cover)'{-c,--cover}'[Cover media file]:file:_files' \
                        '(-o --output)'{-o,--output}'[Output file path]:file:_files' \
                        '(-t --technique)'{-t,--technique}"[Steganography technique]:technique:{_describe 'technique' techniques}" \
                        '(-p --password)'{-p,--password}'[Password for encryption]:password:' \
                        '(-r --redundancy)'{-r,--redundancy}'[Reed-Solomon redundancy (0.0-1.0)]:redundancy:(0.2 0.3 0.5)' \
                        '(-q --quality)'{-q,--quality}'[Quality level (1-100)]:quality:(80 90 95 100)' \
                        '(-j --json)'{-j,--json}'[Output result in JSON format]' \
                        $global_flags
                    ;;
                extract)
                    _arguments \
                        '(-i --input)'{-i,--input}'[Input stego file]:file:_files' \
                        '(-o --output)'{-o,--output}'[Output file path]:file:_files' \
                        '(-p --password)'{-p,--password}'[Password for decryption]:password:' \
                        '(-j --json)'{-j,--json}'[Output result in JSON format]' \
                        $global_flags
                    ;;
                analyze)
                    local -a analyze_commands
                    analyze_commands=(
                        'capacity:Analyze embedding capacity'
                    )
                    _arguments \
                        "1: :{_describe 'analyze command' analyze_commands}" \
                        '(-c --cover)'{-c,--cover}'[Cover media file]:file:_files' \
                        '(-t --technique)'{-t,--technique}"[Steganography technique]:technique:{_describe 'technique' techniques}" \
                        '(-j --json)'{-j,--json}'[Output result in JSON format]' \
                        $global_flags
                    ;;
                formats|version|help)
                    _arguments $global_flags
                    ;;
            esac
            ;;
    esac
}

_shadowforge "$@"

# vim: ft=zsh sw=4 ts=4 et
