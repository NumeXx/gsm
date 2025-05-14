#!/usr/bin/env bash

# Script to create the default GSM configuration directory and an empty config file.
# This script is intended to be run by users if they want to manually ensure
# the configuration directory and a base file exist, or by other setup scripts.

GSM_CONFIG_DIR="$HOME/.gsm"
GSM_CONFIG_FILE="$GSM_CONFIG_DIR/config.json"
GSM_WORDLIST_DIR="$GSM_CONFIG_DIR/wordlist" # Future use for custom wordlists maybe

echo "[GSM Setup] Checking GSM user configuration..."

# Create base GSM config directory
if [ ! -d "$GSM_CONFIG_DIR" ]; then
    echo "[GSM Setup] Creating GSM configuration directory: $GSM_CONFIG_DIR"
    mkdir -p "$GSM_CONFIG_DIR"
    if [ $? -ne 0 ]; then
        echo "[GSM Setup] Error: Failed to create directory $GSM_CONFIG_DIR. Please check permissions." >&2
        exit 1
    fi
else
    echo "[GSM Setup] GSM configuration directory already exists: $GSM_CONFIG_DIR"
fi

# Create wordlist subdirectory (even if not used by go:embed, good for future)
if [ ! -d "$GSM_WORDLIST_DIR" ]; then
    echo "[GSM Setup] Creating wordlist subdirectory: $GSM_WORDLIST_DIR"
    mkdir -p "$GSM_WORDLIST_DIR"
    if [ $? -ne 0 ]; then
        echo "[GSM Setup] Error: Failed to create directory $GSM_WORDLIST_DIR." >&2
        # Not exiting as it's not critical for current go:embed version
    fi
else
    echo "[GSM Setup] Wordlist subdirectory already exists: $GSM_WORDLIST_DIR"
fi


# Create default (empty) config file if it doesn't exist
# The main GSM application will also create this on first run if missing.
if [ ! -f "$GSM_CONFIG_FILE" ]; then
    echo "[GSM Setup] Creating new empty GSM configuration file: $GSM_CONFIG_FILE"
    # Ensure the file is created with reasonably secure permissions (read/write for user only)
    echo "{
  \"connections\": []
}" > "$GSM_CONFIG_FILE"
    chmod 600 "$GSM_CONFIG_FILE" # Set permissions
    if [ $? -ne 0 ]; then
        echo "[GSM Setup] Error: Failed to create or set permissions for $GSM_CONFIG_FILE." >&2
        exit 1
    fi
    echo "[GSM Setup] Empty GSM config file created successfully."
else
    echo "[GSM Setup] GSM configuration file already exists: $GSM_CONFIG_FILE"
fi

echo "[GSM Setup] Configuration check complete." 