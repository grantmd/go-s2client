#!/bin/bash

pushd 3.16.1
docker build . -t grantmd/sc2client:3.16.1
popd

pushd 3.17
docker build . -t grantmd/sc2client:3.17
popd

pushd 4.0.2
docker build . -t grantmd/sc2client:4.0.2
popd

pushd 4.1.2
docker build . -t grantmd/sc2client:4.1.2 -t grantmd/sc2client:latest
popd
