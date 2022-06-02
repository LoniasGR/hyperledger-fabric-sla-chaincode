pushd ./docker/

if [ "$1" = "down" ]; then
docker-compose down
else
docker-compose down
docker-compose up -d
fi
popd
