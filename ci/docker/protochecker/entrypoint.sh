#!/bin/bash
set -euo pipefail

# Configure git to trust the workspace directory
git config --global --add safe.directory "${WORKSPACE_DIR}"

echo "=== Checking for protobuf file changes ==="
echo ""

# Track overall status
EXIT_CODE=0

# Check GUI protobuf files
# This is an ugly way to have comparison since GUI's protobuf generating image has
# a complex way of regenerating files, which creates many permission related problems
# in the CI job
if [ -d "${GUI_PB_BACKUP_DIR}" ]; then
	echo "Comparing generated GUI files with committed versions..."

	# Compare files excluding google directories (symlinks)
	BACKUP_FILES=$(cd "${GUI_PB_BACKUP_DIR}" && find . -type f ! -path '*/google/*' | sort)
	GENERATED_FILES=$(cd "${GUI_PB_DIR}" && find . -type f ! -path '*/google/*' | sort)

	# First check if the file lists match (let's us skip file-by-file comparison)
	if [ "$BACKUP_FILES" != "$GENERATED_FILES" ]; then
		echo "ERROR: Generated GUI protobuf files differ from committed versions!"
		echo "File list mismatch detected."
		EXIT_CODE=1
	else
		# Compare each file's content
		HAS_DIFF=0
		for file in $BACKUP_FILES; do
			if [ -n "$file" ]; then
				if ! diff -q "${GUI_PB_BACKUP_DIR}/$file" "${GUI_PB_DIR}/$file" >/dev/null 2>&1; then
					if [ "$HAS_DIFF" = "0" ]; then
						echo "ERROR: Generated GUI protobuf files differ from committed versions!"
						echo "Differences found:"
						HAS_DIFF=1
					fi
					echo "  $file"
				fi
			fi
		done

		if [ "$HAS_DIFF" = "1" ]; then
			EXIT_CODE=1
		else
			echo "✓ GUI protobuf files match"
		fi
	fi

	rm -rf "${GUI_PB_BACKUP_DIR}"
else
	echo "⚠ No backup found - skipping GUI protobuf check"
fi

echo ""

# Check daemon protobuf files
echo "Comparing generated daemon files with committed versions..."
if ! git diff --quiet --exit-code "${DAEMON_PB_DIR}"; then
	echo "ERROR: Daemon protobuf files changed!"
	echo "Changed files:"
	git diff --name-only "${DAEMON_PB_DIR}"
	EXIT_CODE=1
else
	echo "✓ Daemon protobuf files match"
fi

echo ""

# Final status
if [ "$EXIT_CODE" = "0" ]; then
	echo "SUCCESS: All protobuf files match committed versions."
	exit 0
else
	echo "FAILURE: Protobuf files need to be regenerated."
	echo "To fix this, regenerate protobuf files locally using 'mage generate:protobufDocker', 'rps docker generate protobuf', and commit the changes."
	exit 1
fi
