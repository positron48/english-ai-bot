#!/usr/bin/env python3
import os
import pathlib
import sys
import unittest

SCRIPTS = pathlib.Path(__file__).resolve().parent
sys.path.insert(0, str(SCRIPTS))
os.environ["READING_COVER_PROMPT_MOCK"] = "1"

from reading_cover_prompt import (  # noqa: E402
    _salvage_image_prompt_value,
    build_cover_prompt,
    build_cover_prompt_messages,
    extract_leading_sentences,
    extract_scene_candidates,
    finalize_cover_prompt,
    ensure_no_text_directive,
    parse_image_prompt,
    plain_text_from_doc,
    split_sentences,
)


class ReadingCoverPromptTest(unittest.TestCase):
    def test_plain_text_from_doc(self):
        doc = {
            "title": "Café talk",
            "reading_passage": {
                "segments": [
                    {"text": "Hola", "text_translation_ru": "Привет"},
                    {"text": "¿Qué tal?"},
                ],
            },
        }
        text = plain_text_from_doc(doc)
        self.assertIn("TITLE: Café talk", text)
        self.assertIn("Hola — Привет", text)

    def test_plain_text_for_cover_skips_ru_and_truncates(self):
        doc = {
            "title": "Long",
            "reading_passage": {
                "segments": [
                    {"text": "Hola", "text_translation_ru": "Привет"},
                    *[{"text": f"line {i}"} for i in range(20)],
                ],
            },
        }
        full = plain_text_from_doc(doc)
        short = plain_text_from_doc(doc, for_cover=True)
        self.assertIn("Привет", full)
        self.assertNotIn("Привет", short)
        self.assertGreater(len(full.splitlines()), len(short.splitlines()))

    def test_split_sentences_and_cover_excerpt(self):
        text = "Primera oración. Segunda oración larga. Tercera. Cuarta."
        self.assertEqual(
            split_sentences(text),
            ["Primera oración.", "Segunda oración larga.", "Tercera.", "Cuarta."],
        )
        self.assertEqual(
            extract_leading_sentences(text, max_sentences=2, max_chars=200),
            "Primera oración. Segunda oración larga.",
        )
        doc = {
            "title": "C1 demo",
            "reading_passage": {
                "segments": [
                    {"text": "Una junta revisa el contrato."},
                    {"text": "Los vecinos debaten en la plaza."},
                    {"text": "El ayuntamiento publica el plan."},
                    {"text": "Nadie lee el anexo final."},
                ],
            },
        }
        cover = plain_text_from_doc(doc, for_cover=True)
        self.assertIn("TITLE: C1 demo", cover)
        self.assertIn("Una junta revisa el contrato.", cover)
        self.assertNotIn("anexo final", cover)
        self.assertLessEqual(len(cover), 400)

    def test_build_cover_prompt_messages_fewshot(self):
        msg = build_cover_prompt_messages("TITLE: La letra ñ\n…", "es", "La letra ñ")[0]["content"]
        self.assertIn("Title: La letra ñ", msg)
        self.assertIn("Example:", msg)
        self.assertIn("Do not include any visible text", msg)
        self.assertIn("letters", msg)
        self.assertIn("numbers", msg)
        self.assertIn("Output:", msg)
        self.assertLess(len(msg), 1000)

    def test_parse_image_prompt_json(self):
        raw = '```json\n{"image_prompt": "Two friends sharing tapas at a small cafe table"}\n```'
        self.assertEqual(
            parse_image_prompt(raw, "Tapas en España"),
            "Two friends sharing tapas at a small cafe table",
        )

    def test_parse_image_prompt_uses_json_value_as_is(self):
        raw = '{"image_prompt": "We are given a passage about tapas in Spain."}'
        self.assertEqual(
            parse_image_prompt(raw),
            "We are given a passage about tapas in Spain",
        )

    def test_salvage_truncated_json(self):
        raw = '{"image_prompt": "Nadia holds a red race shirt, frowning at Ivana'
        self.assertEqual(
            _salvage_image_prompt_value(raw),
            "Nadia holds a red race shirt, frowning at Ivana",
        )
        self.assertEqual(
            parse_image_prompt(raw),
            "Nadia holds a red race shirt, frowning at Ivana",
        )

    def test_parse_image_prompt_rejects_reasoning_without_json(self):
        raw = (
            "Okay, let's tackle this. The user wants a JSON object with an image_prompt. "
            "The passage is about Nadia and Ivan."
        )
        with self.assertRaises(ValueError):
            parse_image_prompt(raw)

    def test_finalize_cover_prompt_style(self):
        out = finalize_cover_prompt("a family cooking dinner in a bright kitchen")
        self.assertIn("casual watercolor", out.lower())
        self.assertIn("bright kitchen", out)
        self.assertIn("no text", out.lower())
        self.assertIn("no letters", out.lower())

    def test_ensure_no_text_directive(self):
        out = ensure_no_text_directive("a street market")
        self.assertIn("a street market", out)
        self.assertIn("no text", out.lower())
        self.assertEqual(ensure_no_text_directive("a market, no text"), "a market, no text")

    def test_build_cover_prompt_mock(self):
        doc = {
            "title": "Test",
            "target_language": "es",
            "reading_passage": {"segments": [{"text": "Buenos días"}]},
        }
        prompt = build_cover_prompt(pathlib.Path("."), doc)
        self.assertIn("casual watercolor", prompt.lower())
        self.assertIn("cafe", prompt.lower())


if __name__ == "__main__":
    unittest.main()
