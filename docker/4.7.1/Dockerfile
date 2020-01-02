FROM ubuntu:19.10

# Update the image with required build packages
RUN \
  apt-get update && \
  apt-get -y upgrade && \
  apt-get install -y \
    unzip && \
  apt-get clean && \
  rm -rf /var/lib/apt

WORKDIR /SC2
COPY SC2.4.7.1.zip .

# Unzip the game
RUN unzip -Piagreetotheeula SC2.4.7.1.zip
RUN rm SC2.4.7.1.zip

VOLUME ["/maps"]

RUN ln -s /maps/ /SC2/StarCraftII/maps

EXPOSE 12000
ENTRYPOINT [ "/SC2/StarCraftII/Versions/Base70154/SC2_x64", \
    "-listen", \
    "0.0.0.0", \
    "-port", \
    "12000" ]