FROM golang:1.19-buster
WORKDIR /app
COPY . .
RUN rm -rf .git
RUN make install
RUN apt-get update && apt-get install -y \
    ghostscript \
    poppler-utils \
    imagemagick \
    xz-utils \
    wget \
    && rm -rf /var/lib/apt/lists/* \
    wget https://github.com/pdfcpu/pdfcpu/releases/download/v0.4.1/pdfcpu_0.4.1_Linux_x86_64.tar.xz \
    && tar xf pdfcpu_0.4.1_Linux_x86_64.tar.xz \
    && mv pdfcpu_0.4.1_Linux_x86_64/pdfcpu /usr/local/bin \
    && rm pdfcpu_0.4.1_Linux_x86_64.tar.xz \
    && rm -rf pdfcpu_0.4.1_Linux_x86_64 \
    apt-get update && apt-get install -y \
    tesseract-ocr \
    tesseract-ocr-all \
    && rm -rf /var/lib/apt/lists/*
RUN make build  \
    && chmod +x apario-contribution \
    && useradd -m apario
USER apario
CMD ["make", "run"]
