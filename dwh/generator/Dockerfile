FROM python:3.11

WORKDIR /gen
COPY requirements.txt .
RUN pip install -r requirements.txt
COPY gen.sh .
COPY main.py .
