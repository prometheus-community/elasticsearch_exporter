echo "
export GITHUB_REPO=jessestuart/elasticsearch_exporter
export GO_REPO=github.com/justwatchcom/elasticsearch_exporter
export GOPATH=/home/circleci/go
export GOROOT=/usr/local/go
export PROJECT_PATH=$GOPATH/src/$GO_REPO
export VERSION=$(curl -s https://api.github.com/repos/justwatchcom/elasticsearch_exporter/releases | jq '.[].tag_name' -r | sort -V -r | head -n1)
export REGISTRY=jessestuart
export IMAGE=elasticsearch_exporter
export IMAGE_ID="${REGISTRY}/${IMAGE}:${VERSION}-${TAG}"
" >>$BASH_ENV
