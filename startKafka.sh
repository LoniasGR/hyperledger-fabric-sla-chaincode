rm -rf /tmp/kafka-logs
rm -rf /tmp/zookeeper

cd ~/kafka_2.13-3.0.0
ZOOKEEPER="bin/zookeeper-server-start.sh config/zookeeper.properties"
KAFKA="bin/kafka-server-start.sh config/server.properties"


(trap 'kill 0' SIGINT; eval ${ZOOKEEPER} & eval ${KAFKA})