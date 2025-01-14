#!/bin/bash

# 获取系统架构
ARCH=$(uname -m)

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

echo "Installation complete."
