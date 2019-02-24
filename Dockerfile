FROM iron/go:dev
WORKDIR /app
# Set an env var that matches your github repo name, replace treeder/dockergo here with your repo name
ENV SRC_DIR=/go/src/github.com/tmilner/monzo-customisation/
ARG AUTH_KEY_ARG
ENV AUTH_KEY=$AUTH_KEY_ARG
# Add the source code:
ADD . $SRC_DIR
# Build it:
RUN cd $SRC_DIR; go build -o monzo-customisation; cp monzo-customisation /app/
ENTRYPOINT ./monzo-customisation $AUTH_KEY
