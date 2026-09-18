import vue from 'eslint-plugin-vue'
import ts from 'typescript-eslint'

export default [
  { ignores: ['dist/**', 'node_modules/**'] },
  ...vue.configs['flat/essential'],
  {
    files: ['**/*.ts', '**/*.vue'],
    languageOptions: { parserOptions: { parser: ts.parser, extraFileExtensions: ['.vue'] } },
    rules: {
      'vue/multi-word-component-names': 'off',
      'no-debugger': 'error',
      'no-constant-condition': ['error', { checkLoops: false }],
    },
  },
  { files: ['**/*.ts'], languageOptions: { parser: ts.parser } },
]
