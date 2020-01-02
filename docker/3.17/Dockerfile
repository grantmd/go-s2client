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
COPY SC2.3.17.zip .

# Unzip the game
RUN unzip -Piagreetotheeula SC2.3.17.zip
RUN rm SC2.3.17.zip

VOLUME ["/maps"]

RUN ln -s /maps/ /SC2/StarCraftII/maps

EXPOSE 12000
ENTRYPOINT [ "/SC2/StarCraftII/Versions/Base56787/SC2_x64", \
    "-listen", \
    "0.0.0.0", \
    "-port", \
    "12000" ]