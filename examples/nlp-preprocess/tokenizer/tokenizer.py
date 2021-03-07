import nltk
import nltk.tokenize as nltk_tokenize
from nltk.stem import WordNetLemmatizer
from nltk.corpus import stopwords, wordnet
from typing import List
from collections import Counter
import itertools
from threading import Lock
import unicodedata
import sys
import string
import datetime

unicode_punctuation = ''.join(list(set(chr(i) for i in range(sys.maxunicode)
                                       if unicodedata.category(chr(i)).startswith('P')) | set(string.punctuation)))

months_filter = list(itertools.chain(*[(datetime.date(2020, i, 1).strftime('%B'),
                                        datetime.date(2020, i, 1).strftime('%b')) for i in range(1,13)]))

day_filter = list(itertools.chain(*[(datetime.date(2020, 1, i).strftime('%A'),
                                     datetime.date(2020, 1, i).strftime('%a')) for i in range(1,8)]))

date_filter = list(map(lambda x: x.lower(), months_filter + day_filter))

common_words = ['say', 'go', 'time', 'make', 'said', 'news']

class Tokenizer:
    def __init__(self):
        pass

    def tokenize(self, text):
        pass

    def tokenize_as_sentences(self, text):
        pass

    def _filter_empty_string(self, token):
        return len(token) > 0

    def _lower(self, token: str):
        if self.lower_case:
            return token.lower()
        return token


class TokenProcessor:
    def __init__(self, tokenizer: Tokenizer):
        self.tokenizer = tokenizer

    def process(self, text):
        pass


class TokenFilter:
    def __init__(self, filter_tokens: List[str]=stopwords.words() + date_filter + common_words):
        self.filter_tokens = Counter(filter_tokens)

    def add_tokens(self, tokens: List[str]):
        self.filter_tokens.extend(tokens)

    # ToDo(KMG): Make lower() optional?
    def filter(self, token: str) -> bool:
        result = token.lower() not in self.filter_tokens
        return result


class Stripper:
    def __init__(self):
        pass

    def strip(self, token: str) -> str:
        pass


class NoopStripper:
    def __init__(self):
        pass

    def strip(self, token: str) -> str:
        return token


class StripCharacters(Stripper):
    def __init__(self, strip_characters: str = unicode_punctuation):
        super().__init__()
        self.strip_characters = strip_characters

    def strip(self, token: str) -> str:
        return token.strip(self.strip_characters)


class NLTKSentenceTokenizer(Tokenizer):
    def __init__(self, language='english'):
        super().__init__()
        self.language = language

    def tokenize(self, text):
        return nltk_tokenize.sent_tokenize(text, self.language)

    def tokenize_as_sentences(self, text):
        return nltk_tokenize.sent_tokenize(text, self.language)


class NLTKWordTokenizer(Tokenizer):
    def __init__(self, language='english', token_filter: TokenFilter=TokenFilter(),
                 strip_characters: Stripper=NoopStripper(), sentence_processor=NLTKSentenceTokenizer(),
                 lower_case=True, min_tokens=3):
        super().__init__()
        self.language = language
        self.token_filter = token_filter
        self.strip_characters = strip_characters
        self.sentence_processor = sentence_processor
        self.lower_case =lower_case
        self.min_tokens = min_tokens


    def tokenize(self, text):
        tokens = itertools.chain(*[nltk.word_tokenize(sent)
                                   for sent in self.sentence_processor.tokenize(text)])
        tokens = map(self._lower, tokens)
        tokens = map(self.strip_characters.strip, tokens)
        tokens = filter(self.token_filter.filter, tokens)
        tokens = filter(self._filter_empty_string, tokens)
        return tokens


    def tokenize_as_sentences(self, text):
        original_sentences = self.sentence_processor.tokenize(text)
        sentences = [nltk.word_tokenize(sent)for sent in original_sentences]

        filtered_sentences = []
        for tokens in sentences:
            tokens = map(self._lower, tokens)
            tokens = map(self.strip_characters.strip, tokens)
            tokens = filter(self.token_filter.filter, tokens)
            tokens = list(filter(self._filter_empty_string, tokens))
            if len(tokens) < self.min_tokens:
                tokens = []
            filtered_sentences.append(tokens)
        return filtered_sentences, original_sentences

class NLTKWordLemmatizer(Tokenizer):
    def __init__(self, language='english', token_filter: TokenFilter=TokenFilter(),
                 strip_characters: Stripper=NoopStripper(), sentence_processor=NLTKSentenceTokenizer(),
                 lower_case=True):
        self.language = language
        # ToDo: Need to map the specified language to the tagger language (ISO 639)
        self.tagger_language = "eng"
        self.token_filter = token_filter
        self.strip_characters = strip_characters
        self.sentence_processor = sentence_processor
        self.lower_case =lower_case
        self.sentence_processor = sentence_processor
        self.lemmatizer = WordNetLemmatizer()
        self.tagger = nltk.pos_tag
        self.lock = Lock()

    def _lemmatize(self, tagged_token):
        token, tag = tagged_token

        #
        # NLTK is not thread safe, so need to hold lock for Lemmatization
        # ref: https://github.com/nltk/nltk/issues/803
        #
        with self.lock:
            lemmatized_token = self.lemmatizer.lemmatize(token, self._get_wordnet_pos(tag))

        return lemmatized_token

    # This uses the default, pretrained POS tagger (Treebank corpus).  Need to translate to
    # the wordnet POS tags
    def _get_wordnet_pos(self, treebank_tag):

        if treebank_tag.startswith('J'):
            return wordnet.ADJ
        elif treebank_tag.startswith('S'):
            return wordnet.ADJ_SAT
        elif treebank_tag.startswith('V'):
            return wordnet.VERB
        elif treebank_tag.startswith('N'):
            return wordnet.NOUN
        elif treebank_tag.startswith('R'):
            return wordnet.ADV
        else:
            # ToDo(KMG): Is this right?  Looks like Treebank has a lot of tags...  many more than wordnet
            # Also, apparently, the default lemmatization pos tag is noun, so this is likely fine for this case.
            return wordnet.NOUN

    def tokenize(self, text):
        # ToDo(KMG): What affect does upper/lower have on lemmatization?  I assume we want to preserve capitalization
        # because it may refer to proper nouns...  Maybe best to do post-processing after we tag (we'll know nouns,
        # etc.)
        # untagged_tokens = map(self._lower, untagged_tokens)
        tagged_tokens = itertools.chain(*[self.tagger(nltk.word_tokenize(sent), None, self.tagger_language)
                                          for sent in self.sentence_processor.tokenize(text)])
        tokens = map(self._lemmatize, tagged_tokens)
        tokens = map(self.strip_characters.strip, tokens)
        tokens = filter(self.token_filter.filter, tokens)
        tokens = filter(self._filter_empty_string, tokens)
        return tokens

    def tokenize_as_sentences(self, text):
        original_sentences = self.sentence_processor.tokenize(text)
        tagged_sentences = [self.tagger(nltk.word_tokenize(sent), None, self.tagger_language)
                            for sent in original_sentences]

        filtered_sentences = []
        for tagged_tokens in tagged_sentences:
            tokens = map(self._lemmatize, tagged_tokens)
            tokens = map(self.strip_characters.strip, tokens)
            tokens = filter(self.token_filter.filter, tokens)
            tokens = filter(self._filter_empty_string, tokens)
            filtered_sentences.append(list(tokens))
        return [filtered_sentences, original_sentences]

