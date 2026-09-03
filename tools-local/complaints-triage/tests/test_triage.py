import contextlib
import io
import json
import os
from pathlib import Path
import sys
import unittest
from unittest.mock import patch

sys.path.insert(0, str(Path(__file__).resolve().parents[1]))
import fetch_reports as fetch
import cluster_reports as cluster
import resolve_all_active as resolve


class TriageTest(unittest.TestCase):
    def test_fetch_does_not_drop_reading_and_paginates_without_course(self):
        reading = {'id': 18, 'source_type': 'reading_text', 'grammar_chapter_id': 'free_es_a1_bus'}
        grammar = {'id': 14, 'source_type': 'grammar_training', 'grammar_chapter_id': 'es.grammar.stress'}
        with patch.object(fetch, 'http_json', side_effect=[{'reports': [reading], 'next_cursor': 18}, {'reports': [grammar]}]) as http:
            reports, mode = fetch.fetch_all_reports('https://example.test', 'test-token', 'es', '')
        self.assertEqual(mode, 'unified')
        self.assertEqual([r['id'] for r in fetch.select_course(reports, 'es')], [18, 14])
        self.assertTrue(all('course=' not in c.args[1] for c in http.call_args_list))
        self.assertIn('cursor=18', http.call_args_list[1].args[1])
        self.assertEqual(fetch.select_course(reports, 'en'), [])

    def test_unknown_course_is_visible_and_metadata_identifies_reading(self):
        reports = [
            {'id': 1, 'source_type': 'reading_text', 'payload': {'text_id': 'custom-uuid', 'category_id': 'es_a1'}},
            {'id': 2, 'source_type': 'new_type'},
        ]
        selected = fetch.select_course(reports, 'en')
        self.assertEqual([r['id'] for r in selected], [2])
        self.assertEqual(selected[0]['triage_course'], 'unknown')
        self.assertEqual(fetch.report_course(reports[0]), 'es')

    def test_distinct_reading_and_grammar_entities_are_not_merged(self):
        a = {'id': 1, 'source_type': 'reading_text', 'grammar_chapter_id': 'free_es_a1_rain'}
        b = {'id': 2, 'source_type': 'reading_text', 'payload': {'text_id': 'free_es_a1_bus'}}
        self.assertNotEqual(cluster.cluster_key(a), cluster.cluster_key(b))
        self.assertEqual(cluster.cluster_key(a), cluster.cluster_key({**a, 'id': 3}))
        for kind in ['grammar_chapter', 'grammar_test']:
            self.assertNotEqual(cluster.cluster_key({**a, 'source_type': kind}), cluster.cluster_key({**b, 'source_type': kind, 'grammar_chapter_id': 'es.other'}))
        self.assertNotEqual(cluster.cluster_key({'id': 1, 'source_type': 'future'}), cluster.cluster_key({'id': 2, 'source_type': 'future'}))

    def test_resolve_is_preview_and_uses_matching_token(self):
        env = {'COMPLAINTS_SERVICE_URL_EN': 'https://example.test', 'COMPLAINTS_SERVICE_TOKEN_EN': 'en-test', 'COMPLAINTS_SERVICE_TOKEN_ES': 'es-test'}
        argv = ['resolve_all_active.py', 'en', '--report-ids', '1', '--reason', 'Verified current content']
        with patch.dict(os.environ, env), patch.object(sys, 'argv', argv), patch.object(resolve, 'fetch_all', return_value=[{'id': 1}, {'id': 2}]) as get, patch.object(resolve, 'http') as post, contextlib.redirect_stdout(io.StringIO()) as output:
            self.assertEqual(resolve.main(), 0)
        get.assert_called_once_with('https://example.test', 'en-test', 'en')
        post.assert_not_called()
        self.assertEqual(json.loads(output.getvalue())['report_ids'], [1])

    def test_resolve_rejects_ids_outside_current_snapshot(self):
        env = {'COMPLAINTS_SERVICE_URL_ES': 'https://example.test', 'COMPLAINTS_SERVICE_TOKEN_ES': 'es-test'}
        argv = ['resolve_all_active.py', 'es', '--report-ids', '99', '--reason', 'Verified', '--apply']
        with patch.dict(os.environ, env), patch.object(sys, 'argv', argv), patch.object(resolve, 'fetch_all', return_value=[{'id': 18}]), patch.object(resolve, 'http') as post, contextlib.redirect_stderr(io.StringIO()):
            with self.assertRaises(SystemExit):
                resolve.main()
        post.assert_not_called()


if __name__ == '__main__':
    unittest.main()
