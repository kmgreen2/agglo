# Text Tokenization and Lemmatization Example

echo '{"text": "foo bar baz"}' | docker run -i tokenizer python3 main.py --tokenize_type regular

ToDo: Add pipelines, processes and explaination

Pipeline: Spawn (tokenizer) -> Tee (Original payload to S3) -> Tee (tokens to HTTP endpoint)

