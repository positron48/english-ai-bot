import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createI18n } from 'vue-i18n'
import Panel from './VerbFormsTrainingPanel.vue'
import { apiClient } from '../api/client'
import ru from '../locales/ru.json'
vi.mock('../api/client', () => ({ apiClient: { request: vi.fn() } }))
const fixture = () => ({ session_id: 1, card_id: 8, card_index: 1, total_cards: 10, input_mode: 'choice', options: ['barréis','barres','barre','barren'], prompt: { lemma:'barrer',ru_gloss:'подметать',question:'Vosotros ____ el suelo cada mañana.',example_translation:'Вы подметаете пол каждое утро.',tense:'presente',mood:'indicativo',person:'2',number:'plural' },results:[] })
async function setup() { const wrapper=mount(Panel,{global:{plugins:[createI18n({legacy:false,locale:'ru',messages:{ru}})]}});await flushPromises();return wrapper }
describe('verb conjugation practice',()=>{
 beforeEach(()=>{vi.clearAllMocks();vi.mocked(apiClient.request).mockResolvedValue(fixture())})
 it('shows a compact question with no explanation until help or an answer',async()=>{
  const w=await setup();expect(w.findAll('.practice-option')).toHaveLength(4);expect(w.find('.practice-feedback').exists()).toBe(false);expect(w.find('.practice-hint').exists()).toBe(false)
  expect(w.text()).toContain('Вы подметаете пол каждое утро.');expect(w.find('.practice-tense').exists()).toBe(false);expect(w.text()).toContain('barrer — подметать');expect(w.text()).not.toContain('Нейтральный пример');w.unmount()
 })
 it('keeps feedback visible and advances only on explicit Next',async()=>{
  const w=await setup();const result={...fixture(),feedback:{card_id:8,outcome:'incorrect',assisted:false,chosen_option:'barres',correct_answer:'barréis',sentence:'Vosotros barréis el suelo cada mañana.',translation:'Вы подметаете пол каждое утро.',rule:{regular:true,ending:'éis'},person:'2',number:'plural'}}
  vi.mocked(apiClient.request).mockResolvedValue(result)
  await w.findAll('.practice-option')[1].trigger('click');await flushPromises()
  expect(w.get('.practice-answer').text()).toBe('barréis');expect(w.text()).toContain('окончание -éis');expect(apiClient.request).toHaveBeenCalledTimes(2)
  expect(w.findAll('.practice-option').every(x=>x.attributes('disabled')!==undefined)).toBe(true)
  await w.get('.practice-next button').trigger('click');await flushPromises()
  expect(apiClient.request).toHaveBeenLastCalledWith('/api/verb-training/v2/advance',expect.objectContaining({method:'POST'}));w.unmount()
 })
 it('shows the answer in typed mode and preserves text on network failure',async()=>{
  vi.mocked(apiClient.request).mockResolvedValue({...fixture(),input_mode:'typed',options:[]})
  const w=await setup();await w.get('input').setValue('barreis')
  vi.mocked(apiClient.request).mockRejectedValueOnce(new Error('network'))
  await w.get('form').trigger('submit');await flushPromises()
  expect((w.get('input').element as HTMLInputElement).value).toBe('barreis');expect(w.get('[role="alert"]').text()).toContain('Не удалось')
  vi.mocked(apiClient.request).mockResolvedValue({...fixture(),input_mode:'typed',options:[],feedback:{outcome:'incorrect',correct_answer:'barréis',chosen_option:'barreis',accent_only:true}})
  await w.get('form').trigger('submit');await flushPromises();expect(w.get('.practice-answer').text()).toBe('barréis');expect(w.text()).toContain('ударение');w.unmount()
 })
 it('marks assistance on the server before revealing a hint',async()=>{
  const w=await setup();vi.mocked(apiClient.request).mockResolvedValue({...fixture(),assisted:true,prompt:{...fixture().prompt,rule:{regular:true,ending:'éis'}}})
  const hint=w.findAll('.practice-help-actions button')[0];await hint.trigger('click');await flushPromises()
  expect(apiClient.request).toHaveBeenLastCalledWith('/api/verb-training/v2/help',expect.anything());expect(w.get('.practice-hint').text()).toContain('-éis');w.unmount()
 })
 it('opens the current tense in the reference table despite legacy tense names',async()=>{
  const state={...fixture(),prompt:{...fixture().prompt,tense:'imperfecto'}}
  vi.mocked(apiClient.request).mockResolvedValue(state)
  const w=await setup()
  vi.mocked(apiClient.request).mockImplementation(async(path)=>path.includes('forms-by-lemma') ? {forms:[
    {mood:'indicativo',tense:'presente',person:'1',number:'singular',surface_form:'barro'},
    {mood:'indicativo',tense:'preterito_imperfecto',person:'1',number:'singular',surface_form:'barría'},
    {mood:'indicativo',tense:'imperfecto',person:'1',number:'singular',surface_form:'barría'},
  ]} : {...state,assisted:true})
  await w.findAll('.practice-help-actions button')[1].trigger('click');await flushPromises()
  expect((w.get('select').element as HTMLSelectElement).value).toBe('indicativo|imperfecto')
  expect(w.findAll('tbody tr')).toHaveLength(1);expect(w.get('tbody').text()).toContain('barría');expect(w.get('tbody').text()).not.toContain('barro');w.unmount()
 })

})
