/**
 * The process model the diagram draws. Change vocabulary lives in `../change`.
 */
import type { ChangeDescription } from '../change'

export type { ChangeDescription, ChangeDescriptionKind } from '../change'

export type APIFlowchartKind = 'step' | 'branch' | 'response'

export type APIFlowchartEdge = {
  from: string
  to: string
  label?: string
  change?: ChangeDescription
}

export type APIFlowchartNode = {
  id: string
  kind?: APIFlowchartKind | string
  label: string
  summary?: string
  flowchart_breakdown?: APIFlowchart
  change?: ChangeDescription
}

export type APIFlowchart = {
  id: string
  label?: string
  nodes: APIFlowchartNode[]
  edges: APIFlowchartEdge[]
}
