FROM ubuntu:latest

RUN apt-get update && apt-get install -y ca-certificates

RUN mkdir /home/run
COPY ./booking_history_service /home/run/booking_history_service

WORKDIR /home/run

USER $UID
CMD ["./booking_history_service"]