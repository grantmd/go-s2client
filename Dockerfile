# https://hub.docker.com/_/golang/
FROM golang:alpine

RUN apk update && apk add --no-cache wget unzip net-tools htop git

WORKDIR /SC2

RUN wget http://blzdistsc2-a.akamaihd.net/Linux/SC2.4.1.2.60604_2018_05_16.zip
RUN wget http://blzdistsc2-a.akamaihd.net/MapPacks/Ladder2017Season1.zip
RUN wget http://blzdistsc2-a.akamaihd.net/MapPacks/Ladder2017Season2.zip
RUN wget http://blzdistsc2-a.akamaihd.net/MapPacks/Ladder2017Season3_Updated.zip
RUN wget http://blzdistsc2-a.akamaihd.net/MapPacks/Ladder2017Season4.zip
RUN wget http://blzdistsc2-a.akamaihd.net/MapPacks/Ladder2018Season1.zip
RUN wget http://blzdistsc2-a.akamaihd.net/MapPacks/Ladder2018Season2_Updated.zip
RUN wget http://blzdistsc2-a.akamaihd.net/MapPacks/Melee.zip
RUN wget http://blzdistsc2-a.akamaihd.net/ReplayPacks/3.16.1-Pack_1-fix.zip
RUN wget http://blzdistsc2-a.akamaihd.net/ReplayPacks/3.16.1-Pack_2.zip

RUN git clone --recursive https://github.com/Blizzard/s2client-api.git

RUN unzip -Piagreetotheeula SC2.4.1.2.60604_2018_05_16.zip
RUN unzip -Piagreetotheeula -n -d StarCraftII/Maps Ladder2017Season1.zip
RUN unzip -Piagreetotheeula -n -d StarCraftII/Maps Ladder2017Season2.zip
RUN unzip -Piagreetotheeula -n -d StarCraftII/Maps Ladder2017Season3_Updated.zip
RUN unzip -Piagreetotheeula -n -d StarCraftII/Maps Ladder2017Season4.zip
RUN unzip -Piagreetotheeula -n -d StarCraftII/Maps Ladder2018Season1.zip
RUN unzip -Piagreetotheeula -n -d StarCraftII/Maps Ladder2018Season2_Updated.zip
RUN unzip -Piagreetotheeula -n -d StarCraftII/Maps Melee.zip
RUN cp -r s2client-api/maps/* StarCraftII/Maps/

EXPOSE 12000
ENTRYPOINT [ "/SC2/StarCraftII/Versions/Base60321/SC2_x64", \
    "-listen", \
    "0.0.0.0", \
    "-port", \
    "12000" ]