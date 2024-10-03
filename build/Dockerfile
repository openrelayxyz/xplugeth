FROM golang:1.22

RUN apt update && apt install -y python3-pip && pip install sh --break-system-packages

RUN git clone https://github.com/ethereum/go-ethereum /workdir 

WORKDIR /workdir

COPY build.py /build.py

ENTRYPOINT ["/usr/bin/python3", "/build.py"]