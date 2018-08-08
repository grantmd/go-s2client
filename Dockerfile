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
COPY downloads/SC2.4.1.2.60604_2018_05_16.zip .
#COPY downloads/s2client-api/maps ./s2client-api-maps

# Unzip the game
RUN unzip -Piagreetotheeula SC2.4.1.2.60604_2018_05_16.zip

# Unzip all the maps
# TODO: Mount as volume instead
#RUN unzip -Piagreetotheeula -nj Ladder2017Season1.zip -d StarCraftII/Maps
#RUN unzip -Piagreetotheeula -nj Ladder2017Season2.zip -d StarCraftII/Maps
#RUN unzip -Piagreetotheeula -nj Ladder2017Season3_Updated.zip -d StarCraftII/Maps
#RUN unzip -Piagreetotheeula -nj Ladder2017Season4.zip -d StarCraftII/Maps
#RUN unzip -Piagreetotheeula -nj Ladder2018Season1.zip -d StarCraftII/Maps
#RUN unzip -Piagreetotheeula -nj Ladder2018Season2_Updated.zip -d StarCraftII/Maps
#RUN unzip -Piagreetotheeula -nj Melee.zip -d StarCraftII/Maps
#RUN unzip -Piagreetotheeula -nj 3.16.1-Pack_1-fix.zip -d StarCraftII/Maps
##RUN unzip -Piagreetotheeula -nj 3.16.1-Pack_2.zip -d StarCraftII/Maps
#RUN unzip -Piagreetotheeula -nj mini_games.zip -d StarCraftII/Maps

#RUN cp -r s2client-api-maps/*/* StarCraftII/Maps/

# Delete everything
RUN rm SC2.4.1.2.60604_2018_05_16.zip
#RUN rm mini_games.zip
##RUN rm 3.16.1-Pack_2.zip
#RUN rm 3.16.1-Pack_1-fix.zip
#RUN rm Melee.zip
#RUN rm Ladder2018Season2_Updated.zip
#RUN rm Ladder2018Season1.zip
#RUN rm Ladder2017Season4.zip
#RUN rm Ladder2017Season3_Updated.zip
#RUN rm Ladder2017Season2.zip
#RUN rm Ladder2017Season1.zip

VOLUME ["/maps"]

RUN ln -s /maps/ /SC2/StarCraftII/maps

EXPOSE 12000
ENTRYPOINT [ "/SC2/StarCraftII/Versions/Base60321/SC2_x64", \
    "-listen", \
    "0.0.0.0", \
    "-port", \
    "12000" ]