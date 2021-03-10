FROM python:3.8-alpine
WORKDIR /
# nltk_data generated using: python3 -m nltk.downloader stopwords wordnet averaged_perceptron_tagger universal_tagset punkt
ENV NLTK_DATA /nltk_data
COPY . .
RUN unzip nltk_data.zip
RUN pip install pipenv
RUN apk add build-base
RUN pipenv install --system --deploy
