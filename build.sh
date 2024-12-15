#!/usr/bin/env bash

# Colors
if [[ -t 1 ]]; then
    RED='\033[0;31m'
    GREEN='\033[0;32m'
    YELLOW='\033[1;33m'
    NC='\033[0m'
else
    RED=''
    GREEN=''
    YELLOW=''
    NC=''
fi

# Function to print status messages
echo_status() {
    echo -e "${YELLOW}==> ${NC}$1"
}

echo_success() {
    echo -e "${GREEN}==> SUCCESS: ${NC}$1"
}

echo_error() {
    echo -e "${RED}==> ERROR: ${NC}$1"
}

# Check directory structure
check_structure() {
    if [ ! -d "pkg/git" ] || [ ! -d "cmd/git-multi-push" ]; then
        echo_status "Creating directory structure..."
        mkdir -p pkg/git cmd/git-multi-push
    fi
}

# Build function
build() {
    local target_os=$1
    local target_arch=$2
    local output_name="git-multi-push"
    
    # Add .exe extension for Windows
    if [ "$target_os" = "windows" ]; then
        output_name="${output_name}.exe"
    fi
    
    echo_status "Building for ${target_os}/${target_arch}..."
    
    # Create build directory if it doesn't exist
    mkdir -p build
    
    # Build using the cmd/git-multi-push directory
    GOOS=$target_os GOARCH=$target_arch go build -o "build/${output_name}" ./cmd/git-multi-push
    
    if [ $? -eq 0 ]; then
        echo_success "Built for ${target_os}/${target_arch}"
        # Make binary executable on Unix-like systems
        if [ "$target_os" != "windows" ]; then
            chmod +x "build/${output_name}"
        fi
        return 0
    else
        echo_error "Build failed for ${target_os}/${target_arch}"
        return 1
    fi
}

# Main execution
main() {
    echo_status "Starting build process..."
    
    # Check and create directory structure
    check_structure
    
    # Initialize go module if needed
    if [ ! -f "go.mod" ]; then
        echo_status "Initializing Go module..."
        go mod init git-multi-push
    fi
    
    # Detect OS
    case "$(uname -s)" in
        Linux*)     OS=linux;;
        Darwin*)    OS=darwin;;
        MINGW*)     OS=windows;;
        MSYS*)      OS=windows;;
        *)          OS=linux;;  # Default to linux
    esac
    
    # Build for current platform by default
    if [ $# -eq 0 ]; then
        build "$OS" "amd64"
    else
        case "$1" in
            "windows")
                build "windows" "amd64"
                ;;
            "linux")
                build "linux" "amd64"
                ;;
            "darwin")
                build "darwin" "amd64"
                ;;
            "all")
                build "windows" "amd64" && \
                build "linux" "amd64" && \
                build "darwin" "amd64"
                ;;
            "clean")
                echo_status "Cleaning build directory..."
                rm -rf build
                echo_success "Clean completed"
                ;;
            *)
                echo_error "Unknown target: $1"
                echo "Usage: $0 [windows|linux|darwin|all|clean]"
                exit 1
                ;;
        esac
    fi
}

main "$@"