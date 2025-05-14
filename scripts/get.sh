#!/usr/bin/env bash
# 
# Copyright (c) 2025 Leonardo Faoro & authors
# SPDX-License-Identifier: BSD-3-Clause

set -euo pipefail

cleanup() {
    if [[ "${TEMP_FILE:-}" ]]; then rm -f "$TEMP_FILE"; fi
    if [[ "${TEMP_DIR:-}" ]]; then rm -rf "$TEMP_DIR"; fi
}
trap cleanup EXIT

error() {
    echo "error: $1" >&2
    exit 1
}

is_writable() {
    local path temp_check
    path="$1"
    if [[ ! -d "$path" ]]; then return 1; fi
    temp_check=$(mktemp -t install_check_XXXXXX) || error "failed to create temp file"
    if ! mv "$temp_check" "$path/" 2>/dev/null; then
        rm -f "$temp_check"
        return 1
    fi
    rm -f "$path/$(basename "$temp_check")"
    return 0
}

check_permissions() {
    local path
    path="$1"
    TEMP_FILE=$(mktemp -t install_XXXXXX) || error "failed to create temp file"
    if ! mv "$TEMP_FILE" "$path/" 2>/dev/null; then
        echo "warning: no write permission in $path"
        INSTALL_DIR="$HOME/.local/bin"
        mkdir -p "$INSTALL_DIR" || error "failed to create $INSTALL_DIR"
    fi
    rm -f "$path/$(basename "$TEMP_FILE")" 2>/dev/null
}

check_path() {
    local path
    path="$1"
    if [[ ":$PATH:" != *":$path:"* ]]; then
        echo "Warning: $path is not in your PATH"
        case "$SHELL" in
            *bash) echo "Run: echo 'export PATH=\$PATH:$path' >> ~/.bashrc" ;;
            *zsh)  echo "Run: echo 'export PATH=\$PATH:$path' >> ~/.zshrc" ;;
            *)     echo "Add $path to your PATH" ;;
        esac
    fi
}

# Configuration
APP_NAME=gsm
REPO="NumeXx/gsm"
DOWNLOAD_URL="https://github.com/${REPO}/releases/download"

# get latest version
echo "Fetching latest version..."
echo "Making API request to: https://api.github.com/repos/${REPO}/releases/latest"

if ! API_RESPONSE=$(curl -sSL "https://api.github.com/repos/${REPO}/releases/latest" 2>&1); then
    echo "Error: Failed to fetch from GitHub API"
    echo "Debug: Curl response:"
    echo "$API_RESPONSE"
    error "GitHub API request failed"
fi

VERSION=$(sed 's/"tag_name": "//;s/"//' <<< "$(grep -o '"tag_name": "[^"]*"' <<< "$API_RESPONSE")")
if [[ -z "$VERSION" ]]; then
    echo "Debug: Raw API response:"
    echo "$API_RESPONSE"
    error "failed to determine latest version"
fi
echo "Found version: ${VERSION}"

OS=$(tr '[:upper:]' '[:lower:]' <<< "$(uname -s)")
ARCH=$(uname -m)
case "${ARCH}" in
    x86_64|amd64) ARCH="x86_64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) error "Unsupported architecture: ${ARCH}" ;;
esac
case "${OS}" in
    linux|freebsd|netbsd|openbsd|solaris)
        ARCHIVE_NAME="${APP_NAME}_${VERSION}_${OS}_${ARCH}.tar.gz"
        if is_writable "/usr/local/bin"; then
            INSTALL_DIR="/usr/local/bin"
        else
            INSTALL_DIR="$HOME/.local/bin"
        fi
        ;;
    darwin)
        ARCHIVE_NAME="${APP_NAME}_darwin_all.tar.gz"
        if is_writable "/usr/local/bin"; then
            INSTALL_DIR="/usr/local/bin"
        else
            INSTALL_DIR="$HOME/.local/bin"
        fi
        ;;
    *) error "Unsupported operating system: ${OS}" ;;
esac

# create installation directory
mkdir -p "${INSTALL_DIR}" || error "failed to create installation directory"

# only check permissions if we're not already in a fallback directory
if [[ "$INSTALL_DIR" != "/tmp" && "$INSTALL_DIR" != "$HOME/.local/bin" && "$INSTALL_DIR" != "$HOME/bin" ]]; then
    check_permissions "$INSTALL_DIR"
fi

# download and install binary
DOWNLOAD_ARCHIVE_URL="${DOWNLOAD_URL}/${VERSION}/${ARCHIVE_NAME}"
echo "Downloading ${APP_NAME} ${VERSION} for ${OS}/${ARCH}..."
echo "Attempting to download from: ${DOWNLOAD_ARCHIVE_URL}"

# verify the download url exists before attempting to download
echo "Verifying download URL..."
HTTP_STATUS=$(curl -L -s -o /dev/null -w "%{http_code}" "${DOWNLOAD_ARCHIVE_URL}")
if [[ "$HTTP_STATUS" != "200" ]]; then
    echo "Error: HTTP status code: ${HTTP_STATUS}"
    echo "Debug: Attempting to list available assets..."
    sed 's/"browser_download_url": "//;s/"//' <<< "$(grep -o '"browser_download_url": "[^"]*"' <<< "$API_RESPONSE")"
    error "download URL not accessible: ${DOWNLOAD_ARCHIVE_URL}"
fi

# Create temporary directory for extraction
TEMP_DIR=$(mktemp -d) || error "failed to create temporary directory"

# Download and extract the archive
echo "Downloading binary..."
if ! /usr/bin/curl -fsSL "${DOWNLOAD_ARCHIVE_URL}" -o "${TEMP_DIR}/${ARCHIVE_NAME}" --progress-bar; then
    echo "Error: Failed to download binary"
    error "Download failed"
fi
echo "Extracting binary..."
if ! tar -xzf "${TEMP_DIR}/${ARCHIVE_NAME}" -C "${TEMP_DIR}"; then
    echo "Error: Failed to extract binary"
    error "Failed to extract archive"
fi
echo "Installing binary..."
if ! mv "${TEMP_DIR}/ssm" "${INSTALL_DIR}/${APP_NAME}"; then
    echo "Error: Failed to move binary"
    error "Failed to move binary"
fi
if ! chmod +x "${INSTALL_DIR}/${APP_NAME}"; then
    echo "Error: Failed to set executable permissions"
    error "failed to set executable permissions"
fi
BINARY_PATH="${INSTALL_DIR}/${APP_NAME}"

echo "Successfully installed ${APP_NAME} to: ${BINARY_PATH}"
check_path "${INSTALL_DIR}"

# verify
"${BINARY_PATH}" --version || error "failed to run ${APP_NAME}"
