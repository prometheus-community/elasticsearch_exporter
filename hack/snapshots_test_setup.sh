#!/bin/bash

containerID=$(docker run -d --rm -p 9200:9200 -e discovery.type=single-node -e path.repo=/tmp docker.elastic.co/elasticsearch/elasticsearch:7.5.2)

# Wait for ES to become available
while ! curl http://localhost:9200/_cluster/health >/dev/null 2>&1; do
    sleep 2
done

# Create indices & events
curl -H 'Content-Type: application/json' -XPUT "http://localhost:9200/foo_1/_doc/1" -d '{"title":"abc","content":"hello"}'
curl -H 'Content-Type: application/json' -XPUT "http://localhost:9200/foo_1/_doc/2" -d '{"title":"def","content":"world"}'
curl -H 'Content-Type: application/json' -XPUT "http://localhost:9200/foo_2/_doc/1" -d '{"title":"abc001","content":"hello001"}'
curl -H 'Content-Type: application/json' -XPUT "http://localhost:9200/foo_2/_doc/2" -d '{"title":"def002","content":"world002"}'
curl -H 'Content-Type: application/json' -XPUT "http://localhost:9200/foo_2/_doc/3" -d '{"title":"def003","content":"world003"}'

# Create snapshot repositories
curl -H 'Content-Type: application/json' -XPUT "http://localhost:9200/_snapshot/succeed" -d '{"type": "fs","settings":{"location": "/tmp/succeed"}}'
curl -H 'Content-Type: application/json' -XPUT "http://localhost:9200/_snapshot/fail" -d '{"type": "fs","settings":{"location": "/tmp/fail"}}'

# Create snapshots
curl -XPUT "http://localhost:9200/_snapshot/succeed/visible?wait_for_completion=true"
curl -XPUT "http://localhost:9200/_snapshot/fail/notvisible?wait_for_completion=true"

# Remove access to the "fail" repo
docker exec -it "$containerID" chmod 000 /tmp/fail

## Echo out test cases
## shellcheck disable=SC2016
#echo '{`'"$(curl http://localhost:9200/_snapshot/succeed)"'`, `'"$(curl http://localhost:9200/_snapshot/succeed/_all)"'`},'
## shellcheck disable=SC2016
#echo '{`'"$(curl http://localhost:9200/_snapshot/fail)"'`, `'"$(curl http://localhost:9200/_snapshot/fail/_all)"'`},'
#
## Kill docker container
#docker kill "$containerID"
