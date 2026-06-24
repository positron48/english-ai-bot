#!/usr/bin/env python3
import os
import pathlib
import sys
import unittest

SCRIPTS = pathlib.Path(__file__).resolve().parent
sys.path.insert(0, str(SCRIPTS))
os.environ["READING_COVER_PROMPT_MOCK"] = "1"

from reading_cover_prompt import (  # noqa: E402
    build_cover_prompt,
    parse_image_prompt,
    plain_text_from_doc,
)


class ReadingCoverPromptTest(unittest.TestCase):
    def test_plain_text_from_doc(self):
        doc = {
            "title": "Café talk",
            "reading_passage": {
                "segments": [{"text": "Hola"}, {"text": "¿Qué tal?"}],
            },
        }
        text = plain_text_from_doc(doc)
        self.assertIn("Café talk", text)
        self.assertIn("Hola", text)

    def test_parse_image_prompt_json(self):
        raw = '```json\n{"image_prompt": "flat cafe scene"}\n```'
        self.assertEqual(parse_image_prompt(raw), "flat cafe scene")

    def test_build_cover_prompt_mock(self):
        doc = {
            "title": "Test",
            "target_language": "es",
            "reading_passage": {"segments": [{"text": "Buenos días"}]},
        }
        prompt = build_cover_prompt(pathlib.Path("."), doc)
        self.assertIn("flat illustration", prompt)


if __name__ == "__main__":
    unittest.main()
