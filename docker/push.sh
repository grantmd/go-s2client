#!/bin/bash

pushd 3.16.1
docker push grantmd/sc2client:3.16.1
popd

pushd 3.17
docker push grantmd/sc2client:3.17
popd

pushd 4.0.2
docker push grantmd/sc2client:4.0.2
popd

pushd 4.1.2
docker push grantmd/sc2client:4.1.2
popd

pushd 4.6
docker push grantmd/sc2client:4.6
popd

pushd 4.6.1
docker push grantmd/sc2client:4.6.1
popd

pushd 4.6.2
docker push grantmd/sc2client:4.6.2
popd

pushd 4.7
docker push grantmd/sc2client:4.7
popd

pushd 4.7.1
docker push grantmd/sc2client:4.7.1
popd

pushd 4.10
docker push grantmd/sc2client:4.10
docker push grantmd/sc2client:latest
popd
