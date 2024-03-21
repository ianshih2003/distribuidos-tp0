FROM ubuntu:latest
RUN apt-get update && apt-get install -y netcat
COPY /netcat/netcat.sh /netcat.sh
CMD ["/netcat.sh"]