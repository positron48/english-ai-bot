import importlib.util
from pathlib import Path
import unittest

spec = importlib.util.spec_from_file_location("grammar_review", Path(__file__).with_name("grammar-review.py"))
review = importlib.util.module_from_spec(spec)
spec.loader.exec_module(review)


class ReviewStatusTest(unittest.TestCase):
    def setUp(self):
        self.unit = dict(source="courses/spanish-grammar/chapters/example/03-questions.json",
                         source_sha256="source", context_sha256="theory", question_ids=["q1", "q2"])
        self.report = review.template(self.unit)

    def completed(self):
        self.report.update(editor="editor-task", verifier="reviewer-task", reviewed_at="2026-09-03",
                           verified_at="2026-09-03", phase="done", verification_note="All questions checked",
                           checks=[dict(command="validator", result="pass", evidence="0 errors")])
        for row in self.report["questions"]:
            row.update(decision="ok", verification="ok")

    def test_no_legacy_completion(self):
        self.assertEqual(review.status(self.unit, None)[0], "pending")

    def test_malformed_report_is_invalid(self):
        self.assertEqual(review.status(self.unit, [])[0], "invalid")
        self.report["questions"] = ["not a review record"]
        self.assertEqual(review.status(self.unit, self.report)[0], "invalid")

    def test_empty_file_still_requires_signoff(self):
        self.unit["question_ids"] = []
        self.report = review.template(self.unit)
        self.assertNotEqual(review.status(self.unit, self.report)[0], "done")
        self.completed()
        self.assertEqual(review.status(self.unit, self.report)[0], "done")

    def test_template_is_not_complete(self):
        self.assertEqual(review.status(self.unit, self.report)[0], "in_review")

    def test_empty_editorial_review_can_await_independent_signoff(self):
        self.unit["question_ids"] = []
        self.report = review.template(self.unit)
        self.report.update(editor="editor-task", reviewed_at="2026-09-04",
                           phase="awaiting_verification")
        self.assertEqual(review.status(self.unit, self.report)[0], "awaiting_verification")
        self.report["phase"] = "done"
        self.assertEqual(review.status(self.unit, self.report)[0], "invalid")

    def test_done_needs_all_evidence(self):
        self.completed()
        self.assertEqual(review.status(self.unit, self.report)[0], "done")
        self.report["checks"] = []
        self.assertEqual(review.status(self.unit, self.report)[0], "invalid")

    def test_done_flag_does_not_replace_reading(self):
        self.completed()
        self.report["questions"][1]["decision"] = "pending"
        self.assertEqual(review.status(self.unit, self.report)[0], "in_review")

    def test_editor_cannot_sign_as_verifier(self):
        self.completed()
        self.report["verifier"] = self.report["editor"]
        self.assertEqual(review.status(self.unit, self.report)[0], "invalid")

    def test_missing_or_duplicate_question_invalidates_coverage(self):
        self.completed()
        self.report["questions"][1]["id"] = "q1"
        self.assertEqual(review.status(self.unit, self.report)[0], "invalid")

    def test_source_and_theory_drift(self):
        self.completed()
        for field in ("source_sha256", "context_sha256"):
            changed = dict(self.unit, **{field: "changed"})
            self.assertEqual(review.status(changed, self.report)[0], "stale")

    def test_fixed_question_needs_note(self):
        self.completed()
        self.report["questions"][0]["decision"] = "fixed"
        self.assertEqual(review.status(self.unit, self.report)[0], "invalid")

    def test_reviewer_can_reject_question(self):
        self.completed()
        self.report["questions"][0]["verification"] = "needs_fix"
        self.assertEqual(review.status(self.unit, self.report)[0], "needs_fix")

    def test_inventory_is_explicitly_outside_placement(self):
        units = review.inventory()
        self.assertTrue(units)
        self.assertTrue(all("placement" not in Path(u["source"]).parts for u in units))
        self.assertEqual({u["bank"] for u in units}, {"chapter", "training", "verbs"})
        self.assertEqual(len({u["source"] for u in units}), len(units))
        self.assertEqual(len({u["report"] for u in units}), len(units))


if __name__ == "__main__":
    unittest.main()
