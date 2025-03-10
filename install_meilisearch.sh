#!/bin/sh

# This script can optionally use a GitHub token to increase your request limit (for example, if using this script in a CI).
# To use a GitHub token, pass it through the GITHUB_PAT environment variable.

# GLOBALS

# Colors
RED='\033[31m'
GREEN='\033[32m'
DEFAULT='\033[0m'

# Project name
PNAME='meilisearch'

# GitHub API address
GITHUB_API='https://api.github.com/repos/meilisearch/meilisearch/releases'
# GitHub Release address
GITHUB_REL='https://github.com/meilisearch/meilisearch/releases/download/'

INSTALL_PATH="/usr/local/bin/meilisearch"

# FUNCTIONS

# Gets the version of the latest stable version of Meilisearch by setting the $latest variable.
# Returns 0 in case of success, 1 otherwise.
get_latest() {
    # temp_file is needed because the grep would start before the download is over
    temp_file=$(mktemp -q /tmp/$PNAME.XXXXXXXXX)
    latest_release="$GITHUB_API/latest"

    if [ $? -ne 0 ]; then
        echo "$0: Can't create temp file."
        fetch_release_failure_usage
        exit 1
    fi

    if [ -z "$GITHUB_PAT" ]; then
        curl -s "$latest_release" > "$temp_file" || return 1
    else
        curl -H "Authorization: token $GITHUB_PAT" -s "$latest_release" > "$temp_file" || return 1
    fi

    latest="$(cat "$temp_file" | grep '"tag_name":' | cut -d ':' -f2 | tr -d '"' | tr -d ',' | tr -d ' ')"

    rm -f "$temp_file"
    return 0
}

# Gets the OS by setting the $os variable.
# Returns 0 in case of success, 1 otherwise.
get_os() {
    os_name=$(uname -s)
    case "$os_name" in
    'Darwin')
        os='macos'
        ;;
    'Linux')
        os='linux'
        ;;
     'MINGW'*)
        os='windows'
        ;;
    *)
        return 1
    esac
    return 0
}

# Gets the architecture by setting the $archi variable.
# Returns 0 in case of success, 1 otherwise.
get_archi() {
    architecture=$(uname -m)
    case "$architecture" in
    'x86_64' | 'amd64' )
        archi='amd64'
        ;;
    'arm64')
        # macOS M1/M2
        if [ $os = 'macos' ]; then
            archi='apple-silicon'
        else
            archi='aarch64'
        fi
        ;;
    'aarch64')
        archi='aarch64'
        ;;
    *)
        return 1
    esac
    return 0
}

success_usage() {
    printf "$GREEN%s\n$DEFAULT" "Meilisearch $latest binary successfully downloaded as '$binary_name' file."
    echo ''
    echo 'Run it:'
    echo "    $ ./$PNAME"
    echo 'Usage:'
    echo "    $ ./$PNAME --help"
}

not_available_failure_usage() {
    printf "$RED%s\n$DEFAULT" 'ERROR: Meilisearch binary is not available for your OS distribution or your architecture yet.'
    echo ''
    echo 'However, you can easily compile the binary from the source files.'
    echo 'Follow the steps at the page ("Source" tab): https://www.meilisearch.com/docs/learn/getting_started/installation'
}

fetch_release_failure_usage() {
    echo ''
    printf "$RED%s\n$DEFAULT" 'ERROR: Impossible to get the latest stable version of Meilisearch.'
    echo 'Please let us know about this issue: https://github.com/meilisearch/meilisearch/issues/new/choose'
    echo ''
    echo 'In the meantime, you can manually download the appropriate binary from the GitHub release assets here: https://github.com/meilisearch/meilisearch/releases/latest'
}

fill_release_variables() {
    # Fill $latest variable.
    if ! get_latest; then
        fetch_release_failure_usage
        exit 1
    fi
    if [ "$latest" = '' ]; then
        fetch_release_failure_usage
        exit 1
     fi
     # Fill $os variable.
     if ! get_os; then
        not_available_failure_usage
        exit 1
     fi
     # Fill $archi variable.
     if ! get_archi; then
        not_available_failure_usage
        exit 1
     fi
}

download_binary() {
    fill_release_variables
    echo "Downloading Meilisearch binary $latest for $os, architecture $archi..."
    case "$os" in
        'windows')
            release_file="$PNAME-$os-$archi.exe"
            binary_name="$PNAME.exe"
            ;;
        *)
            release_file="$PNAME-$os-$archi"
            binary_name="$PNAME"
    esac
    # Fetch the Meilisearch binary.
    curl --fail -OL "$GITHUB_REL/$latest/$release_file"
    if [ $? -ne 0 ]; then
        fetch_release_failure_usage
        exit 1
    fi
    mv "$release_file" "$INSTALL_PATH"
    chmod +x "$INSTALL_PATH"
    success_usage
}

# MAIN

main() {
    download_binary
}
main