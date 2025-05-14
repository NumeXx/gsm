#!/bin/bash
# 
# Copyright (c) 2025 Leonardo Faoro & authors
# SPDX-License-Identifier: BSD-3-Clause

# Development script for GSM with live-reloading.
# Requires inotify-tools to be installed (run 'make dev_setup' or install manually).

APP_NAME="gsm"
APP_PATH="."
CMD_TO_RUN="./bin/${APP_NAME}" # Assumes 'make build' places binary in ./bin/

# Ensure inotifywait is available
if ! command -v inotifywait > /dev/null; then
    echo "Error: inotifywait not found. Please install inotify-tools."
    echo "You might be able to install it by running: make dev_setup"
    exit 1
fi

# Function to kill existing gsm processes
kill_gsm() {
    echo "[DevScript] Attempting to kill existing ${APP_NAME} processes..."
    pkill -f "${CMD_TO_RUN}" 2>/dev/null || true
    # Add any other specific pkill patterns if needed
}

# Function to build and run the app
build_and_run() {
    kill_gsm
    echo "[DevScript] Building ${APP_NAME}..."
    if make build; then # Use the Makefile to build
        echo "[DevScript] Build successful. Starting ${APP_NAME}..."
        # Run in a way that allows TUI to take over the terminal
        eval "${CMD_TO_RUN}" # Use eval if CMD_TO_RUN might contain args for gsm later
    else
        echo "[DevScript] Build failed. Please check errors."
    fi
    echo "[DevScript] ${APP_NAME} process exited or failed to start."
}

# Initial build and run
build_and_run

# Watch for changes and rebuild/rerun
echo "[DevScript] Watching for .go file changes in ${APP_PATH} ..."
while true; do
    # Watch for modify, create, delete, move events on .go files
    # -q for quiet, -r for recursive, -e for events
    # --format '%w%f' gives the path and filename
    inotifywait -q -r -e modify,create,delete,move --format '%w%f' ${APP_PATH} |
    while read -r CHANGED_FILE; do
        if [[ "${CHANGED_FILE}" == *.go ]]; then
            echo "[DevScript] Change detected in ${CHANGED_FILE}"
            build_and_run
            # After build_and_run finishes (app exits), re-establish the watch
            echo "[DevScript] Re-watching for .go file changes..."
            break # Break inner loop to re-enter outer inotifywait loop
        fi
    done
done

# Cleanup on exit (though the loop above is infinite)
trap kill_gsm EXIT 