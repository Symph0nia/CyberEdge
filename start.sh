#!/bin/bash

# 先运行 generate.env.sh
./generate.env.sh

# 然后运行 docker-compose
docker-compose up -d