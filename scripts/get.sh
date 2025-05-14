#!/usr/bin/env bash
# Installer for gsm - GSocket Manager
# Inspired by get.sh from lfaoro/ssm and best practices.

set -euo pipefail

APP_NAME="gsm"
REPO="NumeXx/gsm"

# --- Helper Functions ---

cleanup() {
    if [[ -n "${TEMP_DIR:-}" ]] && [[ -d "${TEMP_DIR:-}" ]]; then rm -rf "$TEMP_DIR"; fi
    if [[ -n "${TEMP_FILE_CHECK:-}" ]] && [[ -f "${TEMP_FILE_CHECK:-}" ]]; then rm -f "$TEMP_FILE_CHECK"; fi
}
trap cleanup EXIT SIGINT SIGTERM

error_exit() {
    echo -e "\033[1;31mError:\033[0m $1" >&2
    exit 1
}

# More robust is_writable check using mv from a temp file
is_writable() {
    local target_dir="$1"
    local test_file_name=".gsm_install_writable_check_$(date +%s%N)"
    
    if [[ "$target_dir" == "$HOME"* ]]; then
        if ! mkdir -p "$target_dir" 2>/dev/null; then 
            return 1
        fi
    elif [[ ! -d "$target_dir" ]]; then
        return 1 
    fi

    TEMP_FILE_CHECK=$(mktemp 2>/dev/null || mktemp -t gsm_check 2>/dev/null) 
    if [[ -z "$TEMP_FILE_CHECK" ]] || [[ ! -f "$TEMP_FILE_CHECK" ]]; then
        return 1
    fi

    if mv "$TEMP_FILE_CHECK" "${target_dir}/${test_file_name}" 2>/dev/null; then
        rm -f "${target_dir}/${test_file_name}" 
        TEMP_FILE_CHECK="" 
        return 0 
    else
        rm -f "$TEMP_FILE_CHECK" 
        TEMP_FILE_CHECK="" 
        return 1 
    fi
}

check_path() {
    local path_to_check="$1"
    if [[ ":$PATH:" != *":${path_to_check}:"* ]]; then
        echo -e "\033[1;33mWarning:\033[0m ${path_to_check} is not in your PATH."
        echo "Please add it to your shell configuration file (e.g., ~/.bashrc, ~/.zshrc)."
        echo "Example: export PATH=\"$PATH:${path_to_check}\""
    fi
}

# --- Main Script ---

echo "[${APP_NAME}] Fetching latest release information from GitHub..."
API_URL="https://api.github.com/repos/${REPO}/releases/latest"

if ! API_RESPONSE=$(curl -sSL --retry 3 --retry-delay 2 "$API_URL" 2>&1); then
    echo "[${APP_NAME}] Debug: Failed to fetch from GitHub API. Response:" >&2
    echo "$API_RESPONSE" >&2
    error_exit "GitHub API request failed. Check network or if repository/releases exist."
fi

VERSION=$(echo "$API_RESPONSE" | grep '"tag_name":' | sed -E 's/.*"tag_name": "([^"]+)".*/\1/')
if [[ -z "$VERSION" ]]; then
    echo "[${APP_NAME}] Debug: Raw API response for version check:" >&2
    echo "$API_RESPONSE" >&2
    error_exit "Failed to determine latest version. No release found or API response format changed."
fi
echo "[${APP_NAME}] Latest version: ${VERSION}"

OS_RAW=$(uname -s)
ARCH_RAW=$(uname -m)
OS=""; ARCH=""

case "${OS_RAW}" in
    Linux) OS="Linux" ;; Darwin) OS="Darwin" ;; FreeBSD) OS="Freebsd" ;;
    NetBSD) OS="Netbsd" ;; OpenBSD) OS="Openbsd" ;; SunOS) OS="Solaris" ;;
    CYGWIN*|MINGW*|MSYS*) OS="Windows" ;; *) error_exit "Unsupported OS: ${OS_RAW}" ;;
esac

case "${ARCH_RAW}" in
    x86_64|amd64) ARCH="x86_64" ;; aarch64|arm64) ARCH="arm64" ;;
    i386|i686) ARCH="386" ;; *) error_exit "Unsupported architecture: ${ARCH_RAW}" ;;
esac

VERSION_NO_V=${VERSION#v}
ARCHIVE_NAME="${APP_NAME}_${VERSION_NO_V}_${OS}_${ARCH}.tar.gz"
if [[ "$OS" == "Windows" ]]; then
    ARCHIVE_NAME="${APP_NAME}_${VERSION_NO_V}_${OS}_${ARCH}.zip"
fi
DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${ARCHIVE_NAME}"

echo "[${APP_NAME}] Target: ${OS}/${ARCH}"
echo "[${APP_NAME}] Download URL: ${DOWNLOAD_URL}"

INSTALL_DIR=""
if [[ -n "${XDG_BIN_HOME:-}" ]] && mkdir -p "${XDG_BIN_HOME}" && is_writable "${XDG_BIN_HOME}"; then INSTALL_DIR="${XDG_BIN_HOME}"
elif mkdir -p "${HOME}/.local/bin" && is_writable "${HOME}/.local/bin"; then INSTALL_DIR="${HOME}/.local/bin"
elif mkdir -p "${HOME}/bin" && is_writable "${HOME}/bin"; then INSTALL_DIR="${HOME}/bin"
fi

if [[ -z "$INSTALL_DIR" ]] && [[ "$OS" != "Windows" ]]; then 
    if is_writable "/usr/local/bin"; then 
        INSTALL_DIR="/usr/local/bin"
    fi
fi

if [[ -z "$INSTALL_DIR" ]]; then
    TEMP_INSTALL_PARENT="$(mktemp -d)" 
    INSTALL_DIR="${TEMP_INSTALL_PARENT}/${APP_NAME}_install"
    mkdir -p "${INSTALL_DIR}" || error_exit "Failed to create temp install directory ${INSTALL_DIR}"
    echo -e "\033[1;33mWarning:\033[0m No standard writable binary directory found."
    echo "${APP_NAME} will be installed to: ${INSTALL_DIR}"
    echo "Please add this to your PATH or move '${INSTALL_DIR}/${APP_NAME}' manually."
else
    echo "[${APP_NAME}] Chosen install directory: ${INSTALL_DIR}"
fi

TEMP_DIR=$(mktemp -d) || error_exit "Failed to create temporary download directory."
echo "[${APP_NAME}] Downloading ${ARCHIVE_NAME} to ${TEMP_DIR}..."

HTTP_STATUS=$(curl -L -s -o /dev/null -w "%{http_code}" "${DOWNLOAD_URL}")
if [[ "$HTTP_STATUS" != "200" ]]; then
    echo -e "\033[1;31mError:\033[0m Download URL inaccessible (HTTP: ${HTTP_STATUS}). URL: ${DOWNLOAD_URL}"
    error_exit "Download failed. Check release assets on GitHub."
fi

if ! curl -fsSL --progress-bar "${DOWNLOAD_URL}" -o "${TEMP_DIR}/${ARCHIVE_NAME}"; then 
    error_exit "Failed to download ${APP_NAME} binary archive."
fi

echo "[${APP_NAME}] Extracting ${APP_NAME}..."
EXTRACTED_BINARY_PATH="${TEMP_DIR}/${APP_NAME}"
if [[ "$OS" == "Windows" ]]; then
    if ! unzip -q "${TEMP_DIR}/${ARCHIVE_NAME}" "${APP_NAME}.exe" -d "${TEMP_DIR}"; then 
        unzip -q "${TEMP_DIR}/${ARCHIVE_NAME}" -d "${TEMP_DIR}" 
    fi
    EXTRACTED_BINARY_PATH="${TEMP_DIR}/${APP_NAME}.exe"
else
    if ! tar -xzf "${TEMP_DIR}/${ARCHIVE_NAME}" -C "${TEMP_DIR}" "${APP_NAME}" 2>/dev/null; then 
        if ! tar -xzf "${TEMP_DIR}/${ARCHIVE_NAME}" -C "${TEMP_DIR}" --strip-components=1 "*/${APP_NAME}" 2>/dev/null ; then
             tar -tzf "${TEMP_DIR}/${ARCHIVE_NAME}" 
             error_exit "Failed to extract ${APP_NAME} from tar.gz. Archive structure may be unexpected."
        fi
    fi
fi

if [[ ! -f "${EXTRACTED_BINARY_PATH}" ]]; then
    ls -l "${TEMP_DIR}"
    error_exit "Binary ${APP_NAME} (or .exe) not found in extracted files."
fi

echo "[${APP_NAME}] Installing to ${INSTALL_DIR}/${APP_NAME}..."
if ! mv "${EXTRACTED_BINARY_PATH}" "${INSTALL_DIR}/${APP_NAME}"; then 
    error_exit "Failed to move ${APP_NAME} to ${INSTALL_DIR}. Check permissions."
fi
if [[ "$OS" == "Windows" ]] && [[ -f "${INSTALL_DIR}/${APP_NAME}.exe" ]] && [[ ! -f "${INSTALL_DIR}/${APP_NAME}" ]]; then
    mv "${INSTALL_DIR}/${APP_NAME}.exe" "${INSTALL_DIR}/${APP_NAME}" 
fi

if [[ "$OS" != "Windows" ]]; then 
    if ! chmod +x "${INSTALL_DIR}/${APP_NAME}"; then
        error_exit "Failed to set executable permissions on ${INSTALL_DIR}/${APP_NAME}."
    fi
fi

echo -e "\033[1;32m[${APP_NAME}] Successfully installed to: ${INSTALL_DIR}/${APP_NAME}\033[0m"
check_path "${INSTALL_DIR}"

echo "[${APP_NAME}] Verifying installation..."
# MODIFIED: Changed from --version flag to version subcommand
if ! "${INSTALL_DIR}/${APP_NAME}" version &>/dev/null; then 
    echo -e "\033[1;33mWarning:\033[0m '${INSTALL_DIR}/${APP_NAME} version' command failed or produced no output." 
else
    # MODIFIED: Changed from --version flag to version subcommand
    VERSION_OUTPUT=$("${INSTALL_DIR}/${APP_NAME}" version) 
    echo "[${APP_NAME}] Verification successful: ${VERSION_OUTPUT}"
fi

if [[ "${INSTALL_DIR}" == "${TEMP_INSTALL_PARENT:-/tmp_no_fallback_path}/${APP_NAME}_install" ]]; then
    echo -e "\033[1;33mReminder:\033[0m ${APP_NAME} was installed to a temporary directory."
    echo "Please move \"${INSTALL_DIR}/${APP_NAME}\" to a permanent location in your PATH."
fi
