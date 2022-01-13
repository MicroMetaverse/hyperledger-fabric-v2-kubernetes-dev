IMG=opendco/hlf-api
docker build -t ${IMG} .
# push to docker img server ,https://hub.docker.com/r/opendco/hlf-api
docker login
docker push ${IMG}