import { describe, expect, it } from 'vitest'
import { createI18n } from 'vue-i18n'
import ru from '../../locales/ru.json'
import en from '../../locales/en.json'
import es from '../../locales/es.json'
import { journeyStepGoal, journeyStepTitle } from './labels'
import { journeyResultLabel } from './resultLabel'
import type { JourneyStep } from './types'
const step:JourneyStep={id:'s',mode:'grammar',title:'Грамматика',goal:'Изучить следующий блок и ответить на вопросы',status:'completed',required:false,completedCount:1,targetCount:1,feedback:null,result:null,resultDetails:{kind:'block',answered:1}}
describe('journey locales',()=>{
 for(const locale of ['ru','en','es'])it(`localizes ${locale} without replacing authored titles`,()=>{
  const i18n=createI18n({legacy:false,locale,messages:{ru,en,es}})
  const t=(key:string,values?:Record<string,unknown>)=>i18n.global.t(key,values||{})
  for(const text of [journeyStepTitle(step,t),journeyStepGoal(step,t),journeyResultLabel(step,t)])expect(text).not.toContain('dailyJourney.')
  expect(journeyStepTitle({...step,mode:'picture',title:'A real title'},t)).toBe('A real title')
  expect(journeyStepTitle({...step,title:'grammar'},t)).toBe(journeyStepTitle(step,t))
  expect(journeyStepGoal({...step,mode:'conversation',goal:'Закончить попытку: до 3 реплик',targetCount:3},t)).toContain('3')
 })
})
