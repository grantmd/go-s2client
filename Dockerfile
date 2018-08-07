FROM ubuntu:16.04

# Update the image with required build packages
RUN \
  sed -i 's/# \(.*multiverse$\)/\1/g' /etc/apt/sources.list && \
  apt-get update && \
  apt-get -y upgrade && \
  apt-get install -y \
    net-tools \
    htop \
    python-minimal \
    software-properties-common \
    wget \
    git \
    unzip

WORKDIR /SC2

# Download everything (individual RUN commands so that if the build is interrupted the steps are cached)
RUN wget -q http://blzdistsc2-a.akamaihd.net/Linux/SC2.4.1.2.60604_2018_05_16.zip
RUN wget -q http://blzdistsc2-a.akamaihd.net/MapPacks/Ladder2017Season1.zip
RUN wget -q http://blzdistsc2-a.akamaihd.net/MapPacks/Ladder2017Season2.zip
RUN wget -q http://blzdistsc2-a.akamaihd.net/MapPacks/Ladder2017Season3_Updated.zip
RUN wget -q http://blzdistsc2-a.akamaihd.net/MapPacks/Ladder2017Season4.zip
RUN wget -q http://blzdistsc2-a.akamaihd.net/MapPacks/Ladder2018Season1.zip
RUN wget -q http://blzdistsc2-a.akamaihd.net/MapPacks/Ladder2018Season2_Updated.zip
RUN wget -q http://blzdistsc2-a.akamaihd.net/MapPacks/Melee.zip
RUN wget -q http://blzdistsc2-a.akamaihd.net/ReplayPacks/3.16.1-Pack_1-fix.zip
#RUN wget -q http://blzdistsc2-a.akamaihd.net/ReplayPacks/3.16.1-Pack_2.zip
RUN wget -q https://github.com/deepmind/pysc2/releases/download/v1.2/mini_games.zip

RUN git clone --recursive https://github.com/Blizzard/s2client-api.git

# Download everything
RUN unzip -Piagreetotheeula SC2.4.1.2.60604_2018_05_16.zip
RUN unzip -Piagreetotheeula -n -d StarCraftII/Maps Ladder2017Season1.zip
RUN unzip -Piagreetotheeula -n -d StarCraftII/Maps Ladder2017Season2.zip
RUN unzip -Piagreetotheeula -n -d StarCraftII/Maps Ladder2017Season3_Updated.zip
RUN unzip -Piagreetotheeula -n -d StarCraftII/Maps Ladder2017Season4.zip
RUN unzip -Piagreetotheeula -n -d StarCraftII/Maps Ladder2018Season1.zip
RUN unzip -Piagreetotheeula -n -d StarCraftII/Maps Ladder2018Season2_Updated.zip
RUN unzip -Piagreetotheeula -n -d StarCraftII/Maps Melee.zip
RUN unzip -Piagreetotheeula -n -d StarCraftII/Maps 3.16.1-Pack_1-fix.zip
#RUN unzip -Piagreetotheeula -n -d StarCraftII/Maps 3.16.1-Pack_2.zip
RUN unzip -Piagreetotheeula -n -d StarCraftII/Maps mini_games.zip

RUN cp -r s2client-api/maps/* StarCraftII/Maps/

# Delete everything
RUN rm SC2.4.1.2.60604_2018_05_16.zip
RUN rm mini_games.zip
#RUN rm 3.16.1-Pack_2.zip
RUN rm 3.16.1-Pack_1-fix.zip
RUN rm Melee.zip
RUN rm Ladder2018Season2_Updated.zip
RUN rm Ladder2018Season1.zip
RUN rm Ladder2017Season4.zip
RUN rm Ladder2017Season3_Updated.zip
RUN rm Ladder2017Season2.zip
RUN rm Ladder2017Season1.zip

EXPOSE 12000
ENTRYPOINT [ "/SC2/StarCraftII/Versions/Base60321/SC2_x64", \
    "-listen", \
    "0.0.0.0", \
    "-port", \
    "12000" ]