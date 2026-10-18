import { describe, expect, it } from 'vitest'
import { journeyResultLabel } from './resultLabel'
import type { JourneyStep } from './types'
const t = (key: string, values?: Record<string, unknown>) => key + ':' + JSON.stringify(values || {})
const base: JourneyStep = { id:'step',mode:'picture',status:'completed',title:'Picture',goal:'',required:false,completedCount:1,targetCount:1,feedback:null,result:null }
describe('daily journey result labels', () => {
 it('distinguishes finished attempts from successful quests and completed chapters', () => {
  expect(journeyResultLabel({...base,resultDetails:{attempt_finished:true,passed:false,turns:3}},t)).toContain('questFailed')
  expect(journeyResultLabel({...base,resultDetails:{attempt_finished:true,passed:true,turns:1}},t)).toContain('questPassed')
  expect(journeyResultLabel({...base,mode:'grammar',resultDetails:{kind:'block',answered:1,chapter_passed:false}},t)).toContain('results.block')
 })
 it('shows real practice counts and no credit for skips or unavailable content', () => {
  expect(journeyResultLabel({...base,mode:'verb_forms',resultDetails:{answered:2,correct:0}},t)).toContain('"correct":0,"total":2')
  expect(journeyResultLabel({...base,status:'skipped',resultDetails:{passed:true}},t)).toContain('results.skipped')
  expect(journeyResultLabel({...base,status:'unavailable'},t)).toContain('results.unavailable')
  expect(journeyResultLabel(base,t)).toContain('results.completed')
 })
})
