#!/bin/bash

pushd 3.16.1
wget --continue http://blzdistsc2-a.akamaihd.net/Linux/SC2.3.16.1.zip
popd

pushd 3.17
wget --continue http://blzdistsc2-a.akamaihd.net/Linux/SC2.3.17.zip
popd

pushd 4.0.2
wget --continue http://blzdistsc2-a.akamaihd.net/Linux/SC2.4.0.2.zip
popd

pushd 4.1.2
wget --continue http://blzdistsc2-a.akamaihd.net/Linux/SC2.4.1.2.60604_2018_05_16.zip
popd

pushd 4.6
wget --continue http://blzdistsc2-a.akamaihd.net/Linux/SC2.4.6.0.67926.zip
popd

pushd 4.6.1
wget --continue http://blzdistsc2-a.akamaihd.net/Linux/SC2.4.6.1.68195.zip
popd

pushd 4.6.2
wget --continue http://blzdistsc2-a.akamaihd.net/Linux/SC2.4.6.2.69232.zip
popd

pushd 4.7
wget --continue http://blzdistsc2-a.akamaihd.net/Linux/SC2.AStar.4.7.zip
popd

pushd 4.7.1
wget --continue http://blzdistsc2-a.akamaihd.net/Linux/SC2.4.7.1.zip
popd

pushd 4.10
wget --continue http://blzdistsc2-a.akamaihd.net/Linux/SC2.4.10.zip
popd

mkdir -p downloads/

pushd downloads

wget --continue http://blzdistsc2-a.akamaihd.net/MapPacks/Ladder2017Season1.zip
wget --continue http://blzdistsc2-a.akamaihd.net/MapPacks/Ladder2017Season2.zip
wget --continue http://blzdistsc2-a.akamaihd.net/MapPacks/Ladder2017Season3_Updated.zip
wget --continue http://blzdistsc2-a.akamaihd.net/MapPacks/Ladder2017Season4.zip
wget --continue http://blzdistsc2-a.akamaihd.net/MapPacks/Ladder2018Season1.zip
wget --continue http://blzdistsc2-a.akamaihd.net/MapPacks/Ladder2018Season2_Updated.zip
wget --continue http://blzdistsc2-a.akamaihd.net/MapPacks/Ladder2018Season3.zip
wget --continue http://blzdistsc2-a.akamaihd.net/MapPacks/Ladder2018Season4.zip
wget --continue http://blzdistsc2-a.akamaihd.net/MapPacks/Ladder2019Season1.zip
wget --continue http://blzdistsc2-a.akamaihd.net/MapPacks/Ladder2019Season2.zip
wget --continue http://blzdistsc2-a.akamaihd.net/MapPacks/Ladder2019Season3.zip
wget --continue http://blzdistsc2-a.akamaihd.net/MapPacks/Melee.zip

wget --continue http://blzdistsc2-a.akamaihd.net/ReplayPacks/3.16.1-Pack_1-fix.zip
wget --continue http://blzdistsc2-a.akamaihd.net/ReplayPacks/3.16.1-Pack_2.zip

wget --continue https://github.com/deepmind/pysc2/releases/download/v1.2/mini_games.zip

if [ ! -d s2client-api ]; then
    git clone --recursive https://github.com/Blizzard/s2client-api.git
else
    pushd s2client-api
    git pull -r
    popd
fi

for i in `ls -1 {Ladder,Melee,mini_games}*.zip`; do
    unzip -Piagreetotheeula -nj -d maps $i
done

#unzip -Piagreetotheeula -nj -d replays 3.16.1-Pack_1-fix.zip
#unzip -Piagreetotheeula -nj -d replays 3.16.1-Pack_2.zip

cp -r s2client-api/maps/*/* maps

popd