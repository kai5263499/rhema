# rhema

[![Go Report Card](https://goreportcard.com/badge/github.com/kai5263499/rhema)](https://goreportcard.com/report/github.com/kai5263499/rhema)

rhema is the Greek word for "utterance" or "thing said". This project is all about taking the content of a URI and turning it into compressed audio for faster (for me anyway) consumption.

## Building

Developing on this project is as easy as:
~~~~bash
# Set some env vars
export BUCKET="my-bucket"
export LOCAL_DEV_PATH="/my/local/dev/path"
export LOCAL_CONTENT_PATH="/local/content/path"
export ELASTICSEARCH_URL="http://localhost:9001"
export GOOGLE_APPLICATION_CREDENTIALS="/local/gcp/json/creds"
# make the builder image
make builder-image

# execute the interactive development shell
make exec-interactive
~~~~

## Usage

This repo contains several related sub-projects that are all avaliable as public Docker images.

### Scrape
Use the scrape image to test scraping a url of its content

~~~~bash
docker run \
--rm \
kai5263499/rhema-scrape "${URI}"
~~~~

### process-url
Use the process-url image to process a URL into an mp3 stored in the `${LOCAL_PATH}` directory

~~~~bash
docker run \
--rm \
-e BUCKET="${BUCKET}" \
-v ${LOCAL_PATH}:/data \
--tmpfs /tmp \
kai5263499/rhema-process-url "${URI}"
~~~~

### bot
Use the bot image to connect to have a bot listen for URLS posted on a slack channel and convert those to audio, stored in a `${LOCAL_PATH}` directory

~~~~bash
docker run \
-it --rm \
-e BUCKET="${BUCKET}" \
-e SLACK_TOKEN="${SLACK_TOKEN}" \
-e LOG_LEVEL="${LOG_LEVEL}" \
-v ${LOCAL_CONTENT_PATH}:/data \
-v ${LOCAL_TMP_PATH}:/tmp \
kai5263499/rhema-bot
~~~~

# Misc

There are a few other optional parameters that can be set.

* `WORDS_PER_MINUTE` - How fast espeak should make its resulting speech
* `ATEMPO` - How much to make audio and video content. This is a floating point decimal where the default is 2.0 or 2x the origional speed

Possible espeak-ng voice options for the optional `ESPEAK_VOICE` parameter include the following from `/usr/lib/x86_64-linux-gnu/espeak-ng-data/voices/!v/`:
```
 Andy          Gene    Mario         Tweaky   f1   f5      john         klatt3   m2   m6       norbert   steph    whisper
 Annie         Gene2   Michael       aunty    f2   iven    kaukovalta   klatt4   m3   m7       quincy    steph2   whisperf
 AnxiousAndy   Jacky  'Mr serious'   boris    f3   iven2   klatt        linda    m4   max      rob       steph3   zac
 Denis         Lee     Storm         croak    f4   iven3   klatt2       m1       m5   michel   robert    travis
```

# Deploy to Kubernetes

Helm3 charts are provided to deploy this service into a Kubernetes cluster

~~~~bash
# Install the mechanism for generating let's encrypt certificates for SSL transport
kubectl create namespace cert-manager
helm install cert-manager jetstack/cert-manager --namespace cert-manager --version v1.2.0 --set installCRDs=true

# Install nginx ingress ingress with a public IP
helm install nginx-ingress ingress-nginx/ingress-nginx --set controller.publishService.enabled=true

# Manual step (sorry) - Set your domain's A record to the public IP exposed above

# Finally, install the Rhema suite of services
helm install rhema ./helm
~~~~
