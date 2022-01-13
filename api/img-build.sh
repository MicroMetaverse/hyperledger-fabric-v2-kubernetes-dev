IMG=opendco/hlf-api:v1
docker build -t ${IMG} .
# push to docker img server ,https://hub.docker.com/r/opendco/hlf-api
docker login
docker push ${IMG}