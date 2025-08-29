#!/usr/bin/env bash
set -e 

if [ "$#" -ne 1 ]; then
    echo "missing path to flags folder"
    exit 1
fi

input_folder=$1
output_folder=assets/images/flags


rm -ri "${output_folder}"
mkdir -p "${output_folder}"

json_file="lib/i18n/en/countries.i18n.json"

# Process each file in the input folder
for file in "$input_folder"/*; do
    # Ensure it's a file
    [ -f "$file" ] || continue

    # Extract file name only (without path)
    filename=$(basename -- "$file")

    # Extract ISO Code from the filename
    iso_code=$(echo "$filename" | sed -n 's/.*ISO name=\(.*\)\.svg/\1/p')

    # Skip files that don't match the pattern
    if [[ -z "$iso_code" ]]; then
        echo "Skipping: $filename (Invalid format)"
        continue
    fi

    # Convert ISO code to lowercase
    iso_code_lower=$(echo "$iso_code" | tr '[:upper:]' '[:lower:]')

    # Copy file to output folder with new name (ISO in lowercase)
    cp "$file" "$output_folder/$iso_code_lower.svg"

    echo "Processed: $filename → $iso_code_lower.svg"
done

echo "✅ Processing complete!"
