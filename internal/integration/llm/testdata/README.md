# LLM Integration Test Cases

This directory contains test cases for LLM prompt integration tests.

## Files

- `llm_word_cases.json` - Test cases for word card generation (vocabulary cards)
- `llm_training_cases.json` - Test cases for training card generation

## Format

Each test case is a JSON object with the following fields:

```json
{
  "word": "string",      // The word/text to test
  "expect": "ok|reject|translate|correction",  // Expected outcome
  "note": "string"       // Optional note explaining the test case
}
```

## Adding Test Cases

To add new test cases, simply edit the corresponding JSON file and add a new object to the array.

### Word Cards (`llm_word_cases.json`)

- Use `expect: "ok"` for words that should generate valid word cards with all required fields (JSON response)
- Use `expect: "reject"` for words that should be rejected (proper nouns, nonsense words, etc.) - returns JSON with error=true
- Use `expect: "translate"` for Russian text that should be translated to English (plain text response, NOT JSON, must contain Latin characters)
- Use `expect: "correction"` for English text that should be corrected (plain text response, NOT JSON, must contain Latin characters)

### Training Cards (`llm_training_cases.json`)

- Use `expect: "ok"` for words that should generate valid training cards with senses and distractors
- Use `expect: "reject"` for words that should be rejected (proper nouns, nonsense words, etc.)

## Running Tests

```bash
# Run word card tests only
make llm-words

# Run training card tests only
make llm-training

# Run all LLM tests
make llm-all

# Or directly with go test
go test -tags=integration -v -run '^TestLLM_WordCards$' ./internal/integration/llm/...
go test -tags=integration -v -run '^TestLLM_TrainingCards$' ./internal/integration/llm/...
```

## Requirements

Tests require the following environment variables:
- `AI_URL` - LLM API endpoint
- `AI_API_KEY` - LLM API key
- `AI_MODEL` - (optional) LLM model name (defaults to config default)

If these are not set, tests will be skipped automatically.
