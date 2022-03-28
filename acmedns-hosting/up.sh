#!/bin/bash
set -eu

source .env

mkdir -p ./acmedns-config
cp ./acmedns-config.cfg ./acmedns-config/config.cfg
sed -i 's/{{ACMEDNS_DOMAIN}}/'${ACMEDNS_DOMAIN}'/g' ./acmedns-config/config.cfg
sed -i 's/{{IP_ADDRESS}}/'${IP_ADDRESS}'/g' ./acmedns-config/config.cfg

docker-compose up --force-recreate -d