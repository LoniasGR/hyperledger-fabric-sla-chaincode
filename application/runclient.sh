
ENV_DAL=`echo $DISCOVERY_AS_LOCALHOST`

if [ "$ENV_DAL" != "true" ]
then
    export DISCOVERY_AS_LOCALHOST=true
fi

if [ "$1" = "dev" ]; then
KAFKA_CONFIG="../kafka-config/consumer.properties.dev"
else
KAFKA_CONFIG="../kafka-config/consumer.properties"
fi

echo "run client..."
echo "Kafka config: ${KAFKA_CONFIG}"
rm -rf ./wallet ./keystore ./keys
mkdir -p ./keys
go run client.go -f ${KAFKA_CONFIG}
