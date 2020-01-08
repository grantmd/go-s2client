FROM ubuntu:19.10 as build

# Update the image with required build packages
RUN \
  apt-get update && \
  apt-get install -y --no-install-recommends \
    unzip && \
  apt-get clean && \
  rm -rf /var/lib/apt

WORKDIR /SC2
COPY SC2.4.6.0.67926.zip .

# Unzip the game
RUN unzip -Piagreetotheeula SC2.4.6.0.67926.zip && \
  rm SC2.4.6.0.67926.zip && \
  apt-get remove -y unzip

FROM ubuntu:19.10
COPY --from=build /SC2 /

VOLUME ["/maps"]

RUN ln -s /maps/ /StarCraftII/maps

EXPOSE 12000
ENTRYPOINT [ "/StarCraftII/Versions/Base67926/SC2_x64", \
    "-listen", \
    "0.0.0.0", \
    "-port", \
    "12000" ]