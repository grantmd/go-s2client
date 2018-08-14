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

# Unzip the game
RUN unzip -Piagreetotheeula SC2.4.1.2.60604_2018_05_16.zip
RUN rm SC2.4.1.2.60604_2018_05_16.zip

VOLUME ["/maps"]

RUN ln -s /maps/ /SC2/StarCraftII/maps

EXPOSE 12000
ENTRYPOINT [ "/SC2/StarCraftII/Versions/Base60321/SC2_x64", \
    "-listen", \
    "0.0.0.0", \
    "-port", \
    "12000" ]