import type { APIFlowchartEdge, APIFlowchartNode } from './model'

export type Point = { x: number; y: number }

export type NodeSize = { width: number; height: number }

export type LaidNode = {
  id: string
  node: APIFlowchartNode
  x: number
  y: number
  width: number
  height: number
  rank: number
}

export type LaidEdge = {
  key: string
  edge: APIFlowchartEdge
  points: Point[]
  labelX?: number
  labelY?: number
}

export type DagreLayout = {
  width: number
  height: number
  nodes: LaidNode[]
  edges: LaidEdge[]
}
