#!/bin/bash

generate_random() {
    openssl rand -base64 32
}

cat > .env << EOF
JWT_SECRET=$(generate_random)
SESSION_SECRET=$(generate_random)
EOF