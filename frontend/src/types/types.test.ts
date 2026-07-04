import { describe, it, expectTypeOf } from 'vitest'
import type { TaskInstance, TaskComment, TaskAttachmentMeta } from './index'

describe('extended types', () => {
  it('TaskInstance has priority', () => {
    expectTypeOf<TaskInstance['priority']>().toEqualTypeOf<'low' | 'medium' | 'high'>()
  })
  it('TaskInstance has is_adhoc', () => {
    expectTypeOf<TaskInstance['is_adhoc']>().toEqualTypeOf<boolean>()
  })
  it('TaskComment exists', () => {
    expectTypeOf<TaskComment['author']>().toEqualTypeOf<string>()
  })
  it('TaskAttachmentMeta exists', () => {
    expectTypeOf<TaskAttachmentMeta['size_bytes']>().toEqualTypeOf<number>()
  })
})
