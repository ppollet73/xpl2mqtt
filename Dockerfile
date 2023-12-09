FROM alpine:latest
RUN apk --update add ca-certificates
COPY ./xpl2mqtt /xpl2mqtt
CMD [ "/xpl2mqtt" ]
