# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

TOP = ../../..
include $(TOP)/build.mk

DOCKER_IMAGE = auction_frontend
APP_SRC = ${PWD}

# Note:
# - docker images are not necessarily rebuild if they exist but are outdated.
#   To force rebuild you have two options
#   - do a 'make clobber' first. This ensures you will have the uptodate images
#     but is a broad and slow brush
#   - to just fore rebuilding an image, call `make` with FORCE_REBUILD defined
DOCKER_BUILD_OPTS ?=
ifdef FORCE_REBUILD
	DOCKER_BUILD_OPTS += --no-cache
endif

build: docker avatars

docker:
	# this image is usually built by docker-compose but we still add a
	# convenience target ere until proxy issues with docker-compose are resolved
	$(DOCKER) build $(DOCKER_BUILD_OPTS) -t $(DOCKER_IMAGE) .

run:
	$(DOCKER) run \
		-it \
		-v ${APP_SRC}:/usr/src/app \
		-v /usr/src/app/node_modules \
		-p 5000:5000 \
		$(DOCKER_IMAGE)

cli:
	$(DOCKER) run \
		-it \
		-v ${APP_SRC}:/usr/src/app \
		$(DOCKER_IMAGE) \
		/bin/sh


AVATARS = auctioneer a-telecom b-net c-mobile
AVATAR_DIR = public/img/users
AVATAR_FILES = $(AVATARS:%=$(AVATAR_DIR)/%.svg)

avatars: $(AVATAR_DIR) $(AVATAR_FILES)

$(AVATAR_DIR):
	mkdir $@

public/img/users/auctioneer.svg:
	wget -O $@ "https://avataaars.io/?avatarStyle=Circle&topType=LongHairStraight&accessoriesType=Blank&hairColor=BrownDark&facialHairType=Blank&clotheType=BlazerShirt&eyeType=Default&eyebrowType=Default&mouthType=Default&skinColor=Light"

public/img/users/a-telecom.svg:
	wget -O $@ "https://avataaars.io/?avatarStyle=Circle&topType=LongHairFrida&accessoriesType=Kurt&hairColor=Red&facialHairType=BeardLight&facialHairColor=BrownDark&clotheType=GraphicShirt&clotheColor=Gray01&graphicType=Skull&eyeType=Wink&eyebrowType=RaisedExcitedNatural&mouthType=Disbelief&skinColor=Brown"

public/img/users/b-net.svg:
	wget -O $@ "https://avataaars.io/?avatarStyle=Circle&topType=ShortHairFrizzle&accessoriesType=Prescription02&hairColor=Black&facialHairType=MoustacheMagnum&facialHairColor=BrownDark&clotheType=BlazerSweater&clotheColor=Black&eyeType=Default&eyebrowType=FlatNatural&mouthType=Default&skinColor=Tanned"

public/img/users/c-mobile.svg:
	wget -O $@ "https://avataaars.io/?avatarStyle=Circle&topType=LongHairMiaWallace&accessoriesType=Sunglasses&hairColor=BlondeGolden&facialHairType=Blank&clotheType=BlazerSweater&eyeType=Surprised&eyebrowType=RaisedExcited&mouthType=Smile&skinColor=Pale"



clobber:
	IMAGE=$$(${DOCKER} images ${DOCKER_IMAGE} -q); \
	if [ ! -z "$${IMAGE}" ]; then ${DOCKER} rmi ${IMAGE}; fi
	# make clobber in demo/ also would take care but better safe than sorry
	-$(RM) $(AVATAR_FILES)
