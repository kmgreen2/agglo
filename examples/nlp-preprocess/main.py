import argparse
from enum import Enum
import sys
import json
from tokenizer.tokenizer import NLTKWordTokenizer, NLTKWordLemmatizer


class TokenizeType(Enum):
    REGULAR = 1
    LEMMATIZE = 2


class Args:
    def __init__(self, tokenize_type: TokenizeType, payload_field: str):
        self.tokenize_type = tokenize_type
        self.payload_field = payload_field


def parse_args() -> Args:
    parser = argparse.ArgumentParser(description='Tokenize text')
    parser.add_argument('--tokenize_type', default='regular', help='tokenization type ("regular" or "lemmatize"')
    parser.add_argument('--payload_field', default='text', help='field name of the text payload')
    args = parser.parse_args(sys.argv[1:])
    if args.tokenize_type == "regular":
        return Args(TokenizeType.REGULAR, args.payload_field)
    elif args.tokenize_type == "lemmatize":
        return Args(TokenizeType.LEMMATIZE, args.payload_field)
    else:
        print(f"Invalid type: {args.tokenize_type}")
        sys.exit(1)


def main():
    args = parse_args()
    in_map = json.load(sys.stdin)

    in_text = in_map.get(args.payload_field)

    tokenizer = None
    if args.tokenize_type == TokenizeType.REGULAR:
        tokenizer = NLTKWordTokenizer()
    elif args.tokenize_type == TokenizeType.LEMMATIZE:
        tokenizer = NLTKWordLemmatizer()

    tokens = []
    if tokenizer:
        tokens = list(tokenizer.tokenize(in_text))

    print(json.dumps({
        "tokens": tokens
    }), end='')

if __name__ == '__main__':
    main()


