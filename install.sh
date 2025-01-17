#!/bin/bash

# 获取系统架构
ARCH=$(uname -m)
OS=$(uname -s)

# 判断系统架构并设置变量
if [[ "$ARCH" == "aarch64" ]] || [[ "$ARCH" == "arm64" ]]; then
    ARCH="arm64"
else
    ARCH="amd64"
fi

# 安装 CyberEdge
echo "Installing CyberEdge..."
wget -q https://github.com/Symph0nia/CyberEdge/releases/download/v1.0.9/cyberedge-linux-${ARCH} -O cyberedge
chmod +x cyberedge

# 安装 httpx
echo "Installing httpx..."
wget -q https://github.com/projectdiscovery/httpx/releases/download/v1.6.9/httpx_1.6.9_linux_${ARCH}.zip -O /tmp/httpx.zip
unzip -q /tmp/httpx.zip -d /tmp
mv /tmp/httpx /usr/local/bin/
rm /tmp/httpx.zip

# 安装 subfinder
echo "Installing subfinder..."
wget -q https://github.com/projectdiscovery/subfinder/releases/download/v2.6.7/subfinder_2.6.7_linux_${ARCH}.zip -O /tmp/subfinder.zip
unzip -q /tmp/subfinder.zip -d /tmp
mv /tmp/subfinder /usr/local/bin/
rm /tmp/subfinder.zip

# 安装 ffuf
echo "Installing ffuf..."
wget -q https://github.com/ffuf/ffuf/releases/download/v2.1.0/ffuf_2.1.0_linux_${ARCH}.tar.gz -O /tmp/ffuf.tar.gz
tar -xf /tmp/ffuf.tar.gz -C /tmp
mv /tmp/ffuf /usr/local/bin/
rm /tmp/ffuf.tar.gz

# 安装 fscan
echo "Installing fscan..."
FSCAN_VERSION="latest"
FSCAN_URL="https://github.com/shadow1ng/fscan/releases/$FSCAN_VERSION/download"

case "$OS" in
Linux)
    case "$ARCH" in
    x86_64) BIN="fscan" ;;
    i386 | i686) BIN="fscan32" ;;
    arm64 | aarch64) BIN="fscan_arm64" ;;
    armv6l) BIN="fscan_armv6" ;;
    armv7l) BIN="fscan_armv7" ;;
    mips) BIN="fscan_mips" ;;
    mips64) BIN="fscan_mips64" ;;
    mipsle) BIN="fscan_mipsle" ;;
    solaris) BIN="fscan_solaris" ;;
    freebsd) BIN="fscan_freebsd" ;;
    freebsd32) BIN="fscan_freebsd32" ;;
    freebsd_arm64) BIN="fscan_freebsd_arm64" ;;
    freebsd_armv6) BIN="fscan_freebsd_armv6" ;;
    freebsd_armv7) BIN="fscan_freebsd_armv7" ;;
    *) echo "Unsupported Linux architecture: $ARCH" && exit 1 ;;
    esac
    ;;
Darwin)
    case "$ARCH" in
    x86_64) BIN="fscan_mac" ;;
    arm64) BIN="fscan_mac_arm64" ;;
    *) echo "Unsupported macOS architecture: $ARCH" && exit 1 ;;
    esac
    ;;
*)
    echo "Unsupported OS: $OS" && exit 1
    ;;
esac

wget -q "$FSCAN_URL/$BIN" -O /tmp/fscan
chmod +x /tmp/fscan
mv /tmp/fscan /usr/local/bin/

# 安装 afrog
echo "Installing afrog..."
AFROG_VERSION="3.1.5"
AFROG_BASE_URL="https://github.com/zan8in/afrog/releases/download/v$AFROG_VERSION"

case "$OS" in
Linux)
    case "$ARCH" in
    x86_64) BIN="afrog_${AFROG_VERSION}_linux_amd64.zip" ;;
    arm64 | aarch64) BIN="afrog_${AFROG_VERSION}_linux_arm64.zip" ;;
    *) echo "Unsupported Linux architecture: $ARCH" && exit 1 ;;
    esac
    ;;
Darwin)
    case "$ARCH" in
    x86_64) BIN="afrog_${AFROG_VERSION}_macOS_amd64.zip" ;;
    arm64) BIN="afrog_${AFROG_VERSION}_macOS_arm64.zip" ;;
    *) echo "Unsupported macOS architecture: $ARCH" && exit 1 ;;
    esac
    ;;
WindowsNT)
    case "$ARCH" in
    x86_64) BIN="afrog_${AFROG_VERSION}_windows_amd64.zip" ;;
    arm64) BIN="afrog_${AFROG_VERSION}_windows_arm64.zip" ;;
    *) echo "Unsupported Windows architecture: $ARCH" && exit 1 ;;
    esac
    ;;
*)
    echo "Unsupported OS: $OS" && exit 1
    ;;
esac

wget -q "$AFROG_BASE_URL/$BIN" -O /tmp/afrog.zip
unzip /tmp/afrog.zip -d /tmp
mv /tmp/afrog /usr/local/bin/
rm -f /tmp/afrog.zip

echo "Installation complete."
