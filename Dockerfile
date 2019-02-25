FROM iron/go:dev

WORKDIR /app

ENV SRC_DIR=/go/src/github.com/tmilner/monzo-customisation/

ARG CLIENT_ID_ARG
ARG CLIENT_SECRET_ARG
ARG REDIRECT_URL_ARG

ENV CLIENT_ID=$CLIENT_ID_ARG
ENV CLIENT_SECRET=$CLIENT_SECRET_ARG
ENV REDIRECT_URL=$REDIRECT_URL_ARG

ADD . $SRC_DIR

RUN cd $SRC_DIR; go build -o monzo-customisation; cp monzo-customisation /app/

ENTRYPOINT ./monzo-customisation $AUTH_KEY
