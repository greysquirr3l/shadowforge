# Fish completion for shadowforge CLI

# Commands
complete -c shadowforge -f -n "__fish_use_subcommand" -a "embed" -d "Embed payload into cover media"
complete -c shadowforge -f -n "__fish_use_subcommand" -a "extract" -d "Extract payload from stego media"
complete -c shadowforge -f -n "__fish_use_subcommand" -a "analyze" -d "Analyze media capacity"
complete -c shadowforge -f -n "__fish_use_subcommand" -a "formats" -d "List supported techniques"
complete -c shadowforge -f -n "__fish_use_subcommand" -a "version" -d "Show version information"
complete -c shadowforge -f -n "__fish_use_subcommand" -a "help" -d "Show help information"

# Global flags
complete -c shadowforge -l help -d "Show help information"
complete -c shadowforge -l verbose -d "Enable verbose logging"
complete -c shadowforge -l debug -d "Enable debug logging"
complete -c shadowforge -l config -r -d "Configuration file path"

# Embed command flags
complete -c shadowforge -n "__fish_seen_subcommand_from embed" -s i -l input -r -d "Input file to embed"
complete -c shadowforge -n "__fish_seen_subcommand_from embed" -s c -l cover -r -d "Cover media file"
complete -c shadowforge -n "__fish_seen_subcommand_from embed" -s o -l output -r -d "Output file path"
complete -c shadowforge -n "__fish_seen_subcommand_from embed" -s t -l technique -r -a "lsb dct phase echo zero-width lsb-audio palette" -d "Steganography technique"
complete -c shadowforge -n "__fish_seen_subcommand_from embed" -s p -l password -r -d "Password for encryption"
complete -c shadowforge -n "__fish_seen_subcommand_from embed" -s r -l redundancy -r -a "0.2 0.3 0.5" -d "Reed-Solomon redundancy"
complete -c shadowforge -n "__fish_seen_subcommand_from embed" -s q -l quality -r -a "80 90 95 100" -d "Quality level"
complete -c shadowforge -n "__fish_seen_subcommand_from embed" -s j -l json -d "Output result in JSON format"

# Extract command flags
complete -c shadowforge -n "__fish_seen_subcommand_from extract" -s i -l input -r -d "Input stego file"
complete -c shadowforge -n "__fish_seen_subcommand_from extract" -s o -l output -r -d "Output file path"
complete -c shadowforge -n "__fish_seen_subcommand_from extract" -s p -l password -r -d "Password for decryption"
complete -c shadowforge -n "__fish_seen_subcommand_from extract" -s j -l json -d "Output result in JSON format"

# Analyze command
complete -c shadowforge -n "__fish_seen_subcommand_from analyze" -f -a "capacity" -d "Analyze embedding capacity"
complete -c shadowforge -n "__fish_seen_subcommand_from analyze" -s c -l cover -r -d "Cover media file"
complete -c shadowforge -n "__fish_seen_subcommand_from analyze" -s t -l technique -r -a "lsb dct phase echo zero-width lsb-audio palette" -d "Steganography technique"
complete -c shadowforge -n "__fish_seen_subcommand_from analyze" -s j -l json -d "Output result in JSON format"

# Alias for sforge
complete -c sforge -w shadowforge

# vim: ft=fish sw=4 ts=4 et
