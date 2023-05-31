FROM golang:1.20.4-buster
WORKDIR /app
COPY . .
RUN rm -rf .git
RUN make install
RUN apt-get update && apt-get install -y \
    ghostscript \
    poppler-utils \
    imagemagick \
    libjpeg62-turbo-dev \
    time \
    xz-utils \
    wget \
    && rm -rf /var/lib/apt/lists/*
RUN wget https://github.com/pdfcpu/pdfcpu/releases/download/v0.4.1/pdfcpu_0.4.1_Linux_x86_64.tar.xz \
    && tar xf pdfcpu_0.4.1_Linux_x86_64.tar.xz \
    && mv pdfcpu_0.4.1_Linux_x86_64/pdfcpu /usr/local/bin \
    && rm pdfcpu_0.4.1_Linux_x86_64.tar.xz \
    && rm -rf pdfcpu_0.4.1_Linux_x86_64
RUN apt-get update && apt-get install -y \
    tesseract-ocr \
    tesseract-ocr-all \
    && rm -rf /var/lib/apt/lists/*
RUN go build -a -race -v -o /app/apario-contribution .  \
    && chmod +x /app/apario-contribution \
    && useradd -m apario \
    && chown -R apario:apario /app
USER apario
CMD ["make", "containered"]
