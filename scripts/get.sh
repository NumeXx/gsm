#!/usr/bin/env bash
# Installer for gsm - GSocket Manager
# Heavily inspired by the get.sh script from lfaoro/ssm

set -euo pipefail

APP_NAME="gsm"
REPO="NumeXx/gsm" # IMPORTANT: Change this to your actual GitHub repo path

# --- Helper Functions (error, cleanup, is_writable, check_path) ---

cleanup() {
    if [[ -n "${TEMP_DIR:-}" ]] && [[ -d "${TEMP_DIR:-}" ]]; then rm -rf "$TEMP_DIR"; fi
}
trap cleanup EXIT

error() {
    echo -e "\033[1;31mError:\033[0m $1" >&2
    exit 1
}

is_writable() {
    local path="$1"
    local temp_check
    if [[ ! -d "$path" ]]; then return 1; fi
    temp_check=$(mktemp "${path}/install_check_XXXXXX") 2>/dev/null
    if [[ -z "$temp_check" ]]; then return 1; fi # Failed to create temp file
    rm -f "$temp_check" # Clean up if successful
    return 0
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

echo "[${APP_NAME}] Fetching latest version information from GitHub..."
API_URL="https://api.github.com/repos/${REPO}/releases/latest"

if ! API_RESPONSE=$(curl -sSL "$API_URL" 2>&1); then
    echo "[${APP_NAME}] Debug: Failed to fetch from GitHub API. Response:" >&2
    echo "$API_RESPONSE" >&2
    error "GitHub API request failed. Check network or if the repository and releases exist."
fi

VERSION=$(echo "$API_RESPONSE" | grep '"tag_name":' | sed -E 's/.*"tag_name": "([^"]+)".*/\1/')

if [[ -z "$VERSION" ]]; then
    echo "[${APP_NAME}] Debug: Raw API response for version check:" >&2
    echo "$API_RESPONSE" >&2
    error "Failed to determine the latest version. No release found or API response format changed."
fi
echo "[${APP_NAME}] Latest version found: ${VERSION}"

# Determine OS and Architecture
OS_RAW=$(uname -s)
ARCH_RAW=$(uname -m)

OS=""
case "${OS_RAW}" in
    Linux)    OS="Linux" ;;    # Match GoReleaser's title case for OS
    Darwin)   OS="Darwin" ;;   # macOS
    FreeBSD)  OS="Freebsd" ;; # Note: GoReleaser uses 'Freebsd', not 'FreeBSD' in title usually
    NetBSD)   OS="Netbsd" ;;  
    OpenBSD)  OS="Openbsd" ;; 
    SunOS)    OS="Solaris" ;; # GoReleaser might output Solaris
    CYGWIN*|MINGW*|MSYS*) OS="Windows" ;; # Handle Windows environments
    *) error "Unsupported operating system: ${OS_RAW}" ;;
esac

ARCH=""
case "${ARCH_RAW}" in
    x86_64|amd64) ARCH="x86_64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    i386|i686)    ARCH="386" ;;
    # Add more specific ARM versions if your GoReleaser config distinguishes them (e.g., armv6, armv7)
    # armv7*) ARCH="armv7" ;;
    # armv6*) ARCH="armv6" ;;
    *) error "Unsupported architecture: ${ARCH_RAW}" ;;
esac

# Construct archive name based on GoReleaser's name_template
# {{ .ProjectName }}_{{- .Version }}_{{- title .Os }}_{{- if eq .Arch "amd64" }}x86_64...{{ end }}
# Note: Version might have 'v' prefix from tag, GoReleaser often strips it for asset name.
# We will try both with and without 'v' if needed, or adjust goreleaser name_template for consistency.
VERSION_NO_V=${VERSION#v} # Remove 'v' prefix if present

ARCHIVE_NAME="${APP_NAME}_${VERSION_NO_V}_${OS}_${ARCH}.tar.gz"
DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${ARCHIVE_NAME}"

echo "[${APP_NAME}] Target OS: ${OS}, Arch: ${ARCH}"
echo "[${APP_NAME}] Constructed download URL: ${DOWNLOAD_URL}"

# Determine installation directory
INSTALL_DIR=""
if is_writable "/usr/local/bin"; then
    INSTALL_DIR="/usr/local/bin"
elif is_writable "$HOME/.local/bin"; then
    INSTALL_DIR="$HOME/.local/bin"
elif is_writable "$HOME/bin"; then
    INSTALL_DIR="$HOME/bin"
else
    # Fallback to a temporary directory if no standard writable bin path is found.
    # User will need to move it manually or adjust their PATH.
    TEMP_INSTALL_DIR="$(mktemp -d)/${APP_NAME}_install"
    mkdir -p "${TEMP_INSTALL_DIR}"
    INSTALL_DIR="${TEMP_INSTALL_DIR}"
    echo -e "\033[1;33mWarning:\033[0m Could not find a writable directory in /usr/local/bin, ~/.local/bin, or ~/bin."
    echo "${APP_NAME} will be installed to a temporary directory: ${INSTALL_DIR}"
    echo "Please move the binary to a directory in your PATH manually after installation."
fi

if [[ "$INSTALL_DIR" != "$(mktemp -d)/${APP_NAME}_install" ]]; then # Don't create if it's the temp fallback
    mkdir -p "${INSTALL_DIR}" || error "Failed to create installation directory: ${INSTALL_DIR}"
fi 

echo "[${APP_NAME}] Installing to: ${INSTALL_DIR}"

# Download and install
TEMP_DIR=$(mktemp -d) || error "Failed to create temporary directory for download."
echo "[${APP_NAME}] Downloading ${ARCHIVE_NAME} to ${TEMP_DIR}..."

# Verify URL before download
HTTP_STATUS=$(curl -L -s -o /dev/null -w "%{http_code}" "${DOWNLOAD_URL}")
if [[ "$HTTP_STATUS" != "200" ]]; then
    echo -e "\033[1;31mError:\033[0m Download URL not accessible (HTTP Status: ${HTTP_STATUS})."
    echo "Attempted URL: ${DOWNLOAD_URL}"
    echo "Please check the release page (${REPO}/releases) for available assets."
    # Try to list available assets from API for debugging
    echo "Available assets for version ${VERSION} from API:"
    echo "$API_RESPONSE" | grep '"browser_download_url":' | sed -E 's/.*"browser_download_url": "([^"]+)".*/  \1/'
    error "Download failed."
fi

if ! curl -fsSL "${DOWNLOAD_URL}" -o "${TEMP_DIR}/${ARCHIVE_NAME}" --progress-bar; then 
    error "Failed to download ${APP_NAME} binary."
fi

echo "[${APP_NAME}] Extracting ${APP_NAME}..."
# Assuming the binary inside the tar.gz is just named ${APP_NAME}
if ! tar -xzf "${TEMP_DIR}/${ARCHIVE_NAME}" -C "${TEMP_DIR}" ${APP_NAME}; then 
    # Fallback: if only the binary is in the archive without a parent dir or if name is different
    # This might happen if goreleaser config changes. For now, assume it's just the binary.
    # A more robust script would list tar contents or try different extraction methods.
    echo "[${APP_NAME}] Standard extraction failed, attempting to extract any single executable named ${APP_NAME}..."
    if ! tar -xzf "${TEMP_DIR}/${ARCHIVE_NAME}" -C "${TEMP_DIR}" --strip-components=1 "*/${APP_NAME}" 2>/dev/null && \
       ! tar -xzf "${TEMP_DIR}/${ARCHIVE_NAME}" -C "${TEMP_DIR}" "${APP_NAME}" 2>/dev/null ; then 
        echo "[${APP_NAME}] Listing contents of ${TEMP_DIR}/${ARCHIVE_NAME}:" 
        tar -tzf "${TEMP_DIR}/${ARCHIVE_NAME}" 
        error "Failed to extract ${APP_NAME} from archive. The structure might be unexpected."
    fi
fi

BINARY_IN_TEMP="${TEMP_DIR}/${APP_NAME}"
if [ ! -f "${BINARY_IN_TEMP}" ]; then # Check if binary was extracted
    echo "[${APP_NAME}] Listing contents of ${TEMP_DIR} after extraction attempt:"
    ls -l "${TEMP_DIR}"
    error "Binary ${APP_NAME} not found in extracted files."
fi

echo "[${APP_NAME}] Installing ${APP_NAME} to ${INSTALL_DIR}/${APP_NAME}..."
if ! mv "${BINARY_IN_TEMP}" "${INSTALL_DIR}/${APP_NAME}"; then
    error "Failed to move ${APP_NAME} to ${INSTALL_DIR}. Check permissions or if the path is a directory."
fi

if ! chmod +x "${INSTALL_DIR}/${APP_NAME}"; then
    error "Failed to set executable permissions on ${INSTALL_DIR}/${APP_NAME}."
fi

echo -e "\033[1;32m[${APP_NAME}] Successfully installed ${APP_NAME} to: ${INSTALL_DIR}/${APP_NAME}\033[0m"
check_path "${INSTALL_DIR}"

# Verify installation
echo "[${APP_NAME}] Verifying installation..."
if ! "${INSTALL_DIR}/${APP_NAME}" --version &>/dev/null; then # Assuming gsm will have a --version flag
    echo -e "\033[1;33mWarning:\033[0m ${APP_NAME} --version command failed or produced no output."
    echo "Installation might be complete, but verification step did not pass as expected."
    echo "Try running '${APP_NAME}' manually."
else
    echo "[${APP_NAME}] Verification successful: $(${INSTALL_DIR}/${APP_NAME} --version)"
fi

if [[ "$INSTALL_DIR" == "$(dirname "${TEMP_INSTALL_DIR:-notemp}")" ]]; then # Check if it was the temp fallback
    echo -e "\033[1;33mReminder:\033[0m ${APP_NAME} was installed to a temporary directory."
    echo "Please move \"${INSTALL_DIR}/${APP_NAME}\" to a permanent location in your PATH."
fi
