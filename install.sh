#!/usr/bin/env bash
#
# cmdpp installer script
# Detects OS and architecture, downloads latest pre-built binary, and installs to ~/.local/bin
#
set -euo pipefail

# ANSI color codes
BOLD="\033[1m"
GREEN="\033[0;32m"
BLUE="\033[0;34m"
YELLOW="\033[1;33m"
RED="\033[0;31m"
RESET="\033[0m"

log_info() {
    printf "${BLUE}[i]${RESET} %s\n" "$1"
}

log_step() {
    printf "${BOLD}==> %s${RESET}\n" "$1"
}

log_success() {
    printf "${GREEN}[✓]${RESET} %s\n" "$1"
}

log_warning() {
    printf "${YELLOW}[!]${RESET} %s\n" "$1"
}

log_error() {
    printf "${RED}[✗] %s${RESET}\n" "$1" >&2
}

REPO="hyuricane/cmdpp"
INSTALL_DIR="${BINDIR:-$HOME/.local/bin}"

printf "\n${BOLD}${GREEN}=====================================${RESET}\n"
printf "${BOLD}   cmdpp Installer (Linux & macOS)   ${RESET}\n"
printf "${BOLD}${GREEN}=====================================${RESET}\n\n"

# 1. Detect Operating System
log_step "Detecting system..."
OS_TYPE="$(uname -s | tr '[:upper:]' '[:lower:]')"
case "${OS_TYPE}" in
    linux*)
        OS="linux"
        OS_DISPLAY="Linux"
        ;;
    darwin*)
        OS="darwin"
        OS_DISPLAY="macOS"
        ;;
    *)
        log_error "Unsupported operating system: $(uname -s)"
        printf "Please visit https://github.com/%s/releases for manual installation.\n" "${REPO}" >&2
        exit 1
        ;;
esac

# 2. Detect Architecture
ARCH_TYPE="$(uname -m | tr '[:upper:]' '[:lower:]')"
case "${ARCH_TYPE}" in
    x86_64|amd64)
        ARCH="amd64"
        ARCH_DISPLAY="x86_64"
        ;;
    arm64|aarch64)
        ARCH="arm64"
        ARCH_DISPLAY="arm64"
        ;;
    *)
        log_error "Unsupported CPU architecture: $(uname -m)"
        printf "Supported architectures: x86_64 (amd64), arm64 (aarch64).\n" >&2
        printf "Please visit https://github.com/%s/releases for alternatives.\n" "${REPO}" >&2
        exit 1
        ;;
esac

log_success "Detected ${OS_DISPLAY} (${ARCH_DISPLAY})"

# 3. Check dependencies (curl or wget, tar)
DOWNLOADER=""
if command -v curl >/dev/null 2>&1; then
    DOWNLOADER="curl"
elif command -v wget >/dev/null 2>&1; then
    DOWNLOADER="wget"
else
    log_error "Neither curl nor wget was found. Please install either tool to continue."
    exit 1
fi

if ! command -v tar >/dev/null 2>&1; then
    log_error "tar command not found. Please install tar to extract the binary archive."
    exit 1
fi

# 4. Determine Version
log_step "Resolving latest release..."
TAG="${CMDPP_VERSION:-}"

if [ -z "${TAG}" ]; then
    if [ "${DOWNLOADER}" = "curl" ]; then
        REDIRECT_URL="$(curl -fsSIL -o /dev/null -w '%{url_effective}' "https://github.com/${REPO}/releases/latest" 2>/dev/null || true)"
        if [ -n "${REDIRECT_URL}" ] && [[ "${REDIRECT_URL}" == *"/tag/"* ]]; then
            TAG="${REDIRECT_URL##*/tag/}"
        fi
    fi
fi

# Fallback to GitHub API if redirect failed or wget used
if [ -z "${TAG}" ]; then
    if [ "${DOWNLOADER}" = "curl" ]; then
        TAG="$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" 2>/dev/null | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/' || true)"
    elif [ "${DOWNLOADER}" = "wget" ]; then
        TAG="$(wget -qO- "https://api.github.com/repos/${REPO}/releases/latest" 2>/dev/null | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/' || true)"
    fi
fi

# Safety fallback default
if [ -z "${TAG}" ]; then
    TAG="v0.1.0"
    log_warning "Could not automatically resolve latest version tag. Falling back to ${TAG}."
else
    log_success "Target release: ${TAG}"
fi

# Clean version (strip leading 'v')
CLEAN_VERSION="${TAG#v}"
ARCHIVE_NAME="cmdpp_${CLEAN_VERSION}_${OS}_${ARCH}.tar.gz"
DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${TAG}/${ARCHIVE_NAME}"

# 5. Create temporary directory
TMP_DIR="$(mktemp -d 2>/dev/null || mktemp -d -t 'cmdpp-install')"
trap 'rm -rf "${TMP_DIR}"' EXIT INT TERM

# 6. Download with visible progress
log_step "Downloading ${ARCHIVE_NAME}..."
log_info "Source: ${DOWNLOAD_URL}"
ARCHIVE_PATH="${TMP_DIR}/${ARCHIVE_NAME}"

if [ "${DOWNLOADER}" = "curl" ]; then
    if ! curl -# -fL -o "${ARCHIVE_PATH}" "${DOWNLOAD_URL}"; then
        log_error "Download failed. Please check your internet connection or verify URL: ${DOWNLOAD_URL}"
        exit 1
    fi
else
    if ! wget --show-progress -q -O "${ARCHIVE_PATH}" "${DOWNLOAD_URL}"; then
        log_error "Download failed. Please check your internet connection or verify URL: ${DOWNLOAD_URL}"
        exit 1
    fi
fi
log_success "Download completed successfully"

# 7. Extract archive
log_step "Extracting binary..."
tar -xzf "${ARCHIVE_PATH}" -C "${TMP_DIR}"
if [ ! -f "${TMP_DIR}/cmdpp" ]; then
    log_error "Extracted archive did not contain the 'cmdpp' executable."
    exit 1
fi
chmod +x "${TMP_DIR}/cmdpp"
log_success "Extracted executable verified"

# 8. Install to destination
log_step "Installing to ${INSTALL_DIR}..."
mkdir -p "${INSTALL_DIR}"
cp -f "${TMP_DIR}/cmdpp" "${INSTALL_DIR}/cmdpp"
chmod +x "${INSTALL_DIR}/cmdpp"
log_success "Installed binary at ${INSTALL_DIR}/cmdpp"

# 9. Check PATH
printf "\n"
if ! echo ":$PATH:" | grep -q ":${INSTALL_DIR}:"; then
    log_warning "${INSTALL_DIR} is not currently in your PATH."
    printf "    To run 'cmdpp' directly, add the following line to your shell profile (~/.bashrc or ~/.zshrc):\n"
    printf "    ${BOLD}export PATH=\"%s:\$PATH\"${RESET}\n\n" "${INSTALL_DIR}"
fi

# 10. Finish
printf "${BOLD}${GREEN}=====================================${RESET}\n"
printf "${BOLD}${GREEN} ✓ cmdpp was installed successfully! ${RESET}\n"
printf "${BOLD}${GREEN}=====================================${RESET}\n"
if command -v "${INSTALL_DIR}/cmdpp" >/dev/null 2>&1; then
    VERSION_OUT="$("${INSTALL_DIR}/cmdpp" --version 2>/dev/null || true)"
    [ -n "${VERSION_OUT}" ] && printf "  %s\n" "${VERSION_OUT}"
fi
printf "\n${BOLD}Quick Start:${RESET}\n"
printf "  • Run ${BOLD}cmdpp${RESET} to open the interactive TUI launcher\n"
printf "  • Run ${BOLD}cmdpp add <name> --cmd \"<command>\"${RESET} to save a command\n"
printf "  • Run ${BOLD}cmdpp --help${RESET} to explore all CLI options\n\n"
