export { SequenceDiagram } from './SequenceDiagram'
export type { SequenceDiagramProps } from './SequenceDiagram'
export type { Sequence, SequenceDiagramModel, SequenceMessage, SequenceParticipant } from './model'
export { toSequenceMermaid } from './sequenceMermaid'
export {
  HIGHLIGHT_CLASS,
  NODE_HIGHLIGHT_CLASS,
  hasHighlights,
  highlightEntries,
  isRemoved,
  stepMarkPlan,
  type HighlightEntry,
  type StepMark,
  type StepMarkPlan,
} from './sequenceHighlights'
export { colorSequenceDescriptions, colorSequenceMarks } from './sequenceDiagramMarks'
export {
  HIGHLIGHT_LABEL,
  HIGHLIGHT_TAG_COLOR,
  changeKind,
  type ChangeDescription,
  type ChangeDescriptionKind,
} from '../change'
