# go-domestia

[![Docker Cloud Build Status](https://img.shields.io/docker/cloud/build/vjacobs/go-domestia.svg)](https://hub.docker.com/r/vjacobs/go-domestia)

Bridges a Domestia DMC-008 controller to Home Assistant using MQTT autodiscovery.

## Usage

1. Create a `domestia.json` configuration file. Example [here](domestia.sample.json)
2. `docker run -v $(pwd)/domestia.json:/domestia.json vjacobs/go-domestia`
