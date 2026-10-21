import type { ChangeDescription } from '../change'

export type SequenceParticipant = {
  id: string
  label: string
  kind?: string
  change?: ChangeDescription
}

export type SequenceMessage = {
  id?: string
  from: string
  to: string
  /** Legacy single label. Used when api and description are both empty. */
  text?: string
  /** Callee method name. */
  api?: string
  /** What this hop does. */
  description?: string
  /** alt | else | end. A branch marker has no from/to. */
  fragment?: 'alt' | 'else' | 'end'
  /** Condition shown on alt or else. */
  guard?: string
  arrow?: string
  change?: ChangeDescription
}

/** The JSON model SequenceDiagram draws. Extra fields on a stored sequence are ignored. */
export type SequenceDiagramModel = {
  autonumber?: boolean
  participants: SequenceParticipant[]
  messages: SequenceMessage[]
}

/** @deprecated use SequenceDiagramModel */
export type Sequence = SequenceDiagramModel
