#!/bin/bash

ESVERSIONS=( "5.4.2" "5.6.16" "6.5.4" "6.8.8" "7.3.0" "7.6.2" )

for ESVERSION in "${ESVERSIONS[@]}"
do
    DOCKERCONTAINERNAME=elasticsearch_exporter_fixtures-${ESVERSION}

    # Start a docker container with the proper version
    echo "Booting container ${DOCKERCONTAINERNAME}"
    docker run --rm -d --name ${DOCKERCONTAINERNAME} -p 9200:9200 -e "discovery.type=single-node" -e "cluster.name=elasticsearch" elasticsearch:${ESVERSION}
    # Wait for the container to boot
    while true
    do
        sleep 1
        curl -fs http://localhost:9200
        res=$?
        if [[ "$res" == "0" ]]
        then
            echo "Container is booted"
            break;
        fi
    done

    # Create the the test indices
    echo "Adding indices"
    curl -s -X PUT http://localhost:9200/twitter > /dev/null
    curl -s -X PUT http://localhost:9200/facebook > /dev/null
    curl -s -X PUT http://localhost:9200/instagram > /dev/null
    curl -s -X PUT http://localhost:9200/viber > /dev/null
    curl -s -X PUT http://localhost:9200/instagram/_settings --header "Content-Type: application/json" -d '
    {
        "index": {
            "blocks": {
            "read_only_allow_delete": "true"
            }
        }
    }' > /dev/null

    curl -s -X PUT http://localhost:9200/twitter/_settings --header "Content-Type: application/json" -d '
    {
        "index": {
            "blocks": {
            "read_only_allow_delete": "true"
            }
        }
    }' > /dev/null

    # Index with some data
    # refresh=wait_for helps to avoid a data race where the documents aren't committed before later curl requests
    curl -s -X PUT http://localhost:9200/foo_1/type1/1?refresh=wait_for --header "Content-Type: application/json" -d '{"title":"abc","content":"hello"}'
    curl -s -X PUT http://localhost:9200/foo_1/type1/2?refresh=wait_for --header "Content-Type: application/json" -d '{"title":"def","content":"world"}'
    curl -s -X PUT http://localhost:9200/foo_2/type1/1?refresh=wait_for --header "Content-Type: application/json" -d '{"title":"abc001","content":"hello001"}'
    curl -s -X PUT http://localhost:9200/foo_2/type1/2?refresh=wait_for --header "Content-Type: application/json" -d '{"title":"def002","content":"world002"}'
    curl -s -X PUT http://localhost:9200/foo_2/type1/3?refresh=wait_for --header "Content-Type: application/json" -d '{"title":"def003","content":"world003"}'

    # Snapshot index
    # echo "Snapshotting index"
    # curl -s -X PUT http://localhost:9200/_snapshot/test1 -d '{"type": "fs","settings":{"location": "/tmp/test1"}}'
    # curl -s -X PUT "http://localhost:9200/_snapshot/test1/snapshot_1?wait_for_completion=true"

    echo "Collecting fixtures"

    # Get cluster health stats
    # Force the task_max_waiting_in_queue_millis to 12 for tests
    curl -s http://localhost:9200/_cluster/health | sed -E 's/"task_max_waiting_in_queue_millis":[[:digit:]]+/"task_max_waiting_in_queue_millis":12/' > fixtures/clusterhealth/${ESVERSION}.json

    # Get index settings
    curl -s http://localhost:9200/_all/_settings > fixtures/indexsettings/${ESVERSION}.json

    # Get index stats
    curl -s http://localhost:9200/_all/_stats > fixtures/indexstats/${ESVERSION}.json

    # Get node stats
    curl -s http://localhost:9200/_nodes/stats > fixtures/nodestats/${ESVERSION}.json

    # Get snapshot stats
    # curl http://localhost:9200/_snapshot/
    # curl -s http://localhost:9200/_snapshot/test1/_all > fixtures/snapshot/${ESVERSION}.json


    # Get cluster settings
    curl -s http://localhost:9200/_cluster/settings/?include_defaults=true > fixtures/clustersettings/${ESVERSION}.json

    # Change some cluster settings
    curl -X PUT http://localhost:9200/_cluster/settings?pretty -H 'Content-Type: application/json' -d'
    {
        "transient": {
            "cluster.routing.allocation.enable": "ALL"
        }
    }
    '

    # Get updated cluster settings
    curl -s http://localhost:9200/_cluster/settings/?include_defaults=true > fixtures/clustersettings/${ESVERSION}-updated.json


    echo "Cleaning up container ${DOCKERCONTAINERNAME}"
    docker stop ${DOCKERCONTAINERNAME}
done
