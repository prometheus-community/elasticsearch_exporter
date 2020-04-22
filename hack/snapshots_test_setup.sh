#!/bin/bash

set -ex

containerID=$(docker run -d --rm -p 9200:9200 -e "discovery.type=single-node" -e "path.repo=/tmp" docker.elastic.co/elasticsearch/elasticsearch:7.5.2)

while ! curl http://localhost:9200/_cluster/health; do
  sleep 2
done

curl -H 'Content-Type: application/json' -XPUT "http://localhost:9200/foo_1/_doc/1" -d '{"title":"abc","content":"hello"}'
curl -H 'Content-Type: application/json' -XPUT "http://localhost:9200/foo_1/_doc/2" -d '{"title":"def","content":"world"}'
curl -H 'Content-Type: application/json' -XPUT "http://localhost:9200/foo_2/_doc/1" -d '{"title":"abc001","content":"hello001"}'
curl -H 'Content-Type: application/json' -XPUT "http://localhost:9200/foo_2/_doc/2" -d '{"title":"def002","content":"world002"}'
curl -H 'Content-Type: application/json' -XPUT "http://localhost:9200/foo_2/_doc/3" -d '{"title":"def003","content":"world003"}'

curl -H 'Content-Type: application/json' -XPUT "http://localhost:9200/_snapshot/succeed" -d '{"type": "fs","settings":{"location": "/tmp/succeed"}}'
curl -H 'Content-Type: application/json' -XPUT "http://localhost:9200/_snapshot/fail" -d '{"type": "fs","settings":{"location": "/tmp/fail"}}'

curl -XPUT "http://localhost:9200/_snapshot/succeed/visible?wait_for_completion=true"
curl -XPUT "http://localhost:9200/_snapshot/fail/notvisible?wait_for_completion=true"

docker exec -it "$containerID" chmod 000 /tmp/fail
