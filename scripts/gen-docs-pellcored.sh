#!/bin/bash

DIR=docs/cli/pellcored

rm -rf $DIR

go install ./cmd/pellcored

pellcored docs --path $DIR

# Recursive function to process files
process_files() {
    local dir="$1"
    
    # Process all files in the directory
    for file in "$dir"/*; do
        if [ -f "$file" ]; then
            # Replace <...> with [...] in the file, otherwise Docusaurus thinks it's a link
            sed -i.bak 's/<\([^<>]*\)>/\[\1\]/g' "$file"

            # Modify the heading by replacing ## pellcored with #
            sed -i.bak 's/^## pellcored /# /g' "$file" 

            # Replace all instances of [appd] with pellcored
            sed -i.bak 's/\[appd\]/pellcored/g' "$file"

            # Remove the pattern (default "SOMETHING")
            sed -i.bak 's/(default ".*")//g' "$file"

            # Remove the last line "###### Auto generated by spf13/cobra on ..."
            sed -i.bak '$ d' "$file"

            # Remove the backup files
            rm -f "$file.bak"
        elif [ -d "$file" ]; then
            # Recurse into subdirectory
            process_files "$file"
        fi
    done
}

# Start processing from the given directory
process_files $DIR
