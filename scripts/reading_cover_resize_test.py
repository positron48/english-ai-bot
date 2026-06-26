#!/usr/bin/env python3
import pathlib
import sys
import tempfile
import unittest

from PIL import Image

SCRIPTS = pathlib.Path(__file__).resolve().parent
sys.path.insert(0, str(SCRIPTS))

from reading_cover_resize import _cover_crop, _cover_fit, resize_cover_assets  # noqa: E402


class ReadingCoverResizeTest(unittest.TestCase):
    def test_cover_fit_preserves_portrait(self):
        img = Image.new("RGB", (800, 1200), color=(10, 20, 30))
        out = _cover_fit(img, (1024, 768))
        self.assertEqual(out.size, (512, 768))

    def test_cover_crop_forces_landscape_thumb(self):
        img = Image.new("RGB", (800, 1200), color=(10, 20, 30))
        out = _cover_crop(img, (400, 300))
        self.assertEqual(out.size, (400, 300))

    def test_resize_cover_assets_hero_keeps_1024x768(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = pathlib.Path(tmp)
            raw = root / "cover_raw.png"
            thumb = root / "cover_thumb.webp"
            hero = root / "cover_hero.webp"
            Image.new("RGB", (1024, 768), color=(255, 0, 0)).save(raw)
            resize_cover_assets(raw, thumb, hero)
            with Image.open(hero) as hero_img:
                self.assertEqual(hero_img.size, (1024, 768))
            with Image.open(thumb) as thumb_img:
                self.assertEqual(thumb_img.size, (400, 300))


if __name__ == "__main__":
    unittest.main()
