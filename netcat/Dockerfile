FROM alpine:latest

ENV MSG="Hola mundo"
ENV SERVER_IP=server
ENV TIMEOUT=3
ENV PORT=12345


COPY /netcat/netcat.sh /netcat.sh
CMD ["sh", "netcat.sh"]