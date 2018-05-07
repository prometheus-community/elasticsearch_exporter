#!/bin/bash

echo 'Determine repo version.'
apk update && apk add curl jq
export VERSION=$(\
  curl -s https://api.github.com/repos/${GITHUB_REPO}/releases/latest | \
  jq -r ".tag_name"
)

echo 'Installing manifest-tool.'
export VERSION=$(cat ~/VERSION)
echo "Downloading manifest-tool."
wget https://github.com/estesp/manifest-tool/releases/download/v0.7.0/manifest-tool-linux-amd64
mv manifest-tool-linux-amd64 /usr/bin/manifest-tool
chmod +x /usr/bin/manifest-tool
manifest-tool --version

echo 'Pushing Docker manifest.'
echo $DOCKERHUB_PASS | docker login -u $DOCKERHUB_USER --password-stdin;

manifest-tool push from-args \
  --platforms linux/arm,linux/arm64,linux/amd64 \
  --template "$REGISTRY/$IMAGE:$VERSION-ARCH" \
  --target "$REGISTRY/$IMAGE:$VERSION"

if [ $CIRCLE_BRANCH == 'master' ]; then
  manifest-tool push from-args \
    --platforms linux/arm,linux/arm64,linux/amd64 \
    --template "$REGISTRY/$IMAGE:$VERSION-ARCH" \
    --target "$REGISTRY/$IMAGE:latest"
fi

echo 'Verifying manifest was persisted remotely.'
manifest-tool inspect "$REGISTRY/$IMAGE:$VERSION"
