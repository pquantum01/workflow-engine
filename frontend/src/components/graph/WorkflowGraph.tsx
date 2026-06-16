import { useCallback, useMemo, useEffect, useState } from 'react'
import ReactFlow, {
  Background,
  Controls,
  MiniMap,
  type Node,
  type Edge,
  MarkerType,
} from 'reactflow'
import 'reactflow/dist/style.css'
import { layoutGraph } from '../../lib/graph'
import type { WorkflowDefinition } from '../../types'

interface Props {
  definition: WorkflowDefinition
  currentStateId?: string
  visitedStates?: string[]
  onNodeClick?: (stateId: string) => void
}

function useIsDark() {
  const [isDark, setIsDark] = useState(() => document.documentElement.classList.contains('dark'))
  useEffect(() => {
    const obs = new MutationObserver(() => {
      setIsDark(document.documentElement.classList.contains('dark'))
    })
    obs.observe(document.documentElement, { attributes: true, attributeFilter: ['class'] })
    return () => obs.disconnect()
  }, [])
  return isDark
}

function nodeStyle(
  isStart: boolean,
  isEnd: boolean,
  isTerminal: boolean,
  isCurrent: boolean,
  isVisited: boolean,
  isDark: boolean
): React.CSSProperties {
  if (isCurrent) {
    return {
      background: isDark ? '#1e1b4b' : '#eef2ff',
      border: '2px solid #6366f1',
      borderRadius: 8,
      color: isDark ? '#c7d2fe' : '#3730a3',
      boxShadow: '0 0 0 3px rgba(99,102,241,0.25)',
    }
  }
  if (isVisited) {
    return {
      background: isDark ? '#27272a' : '#f4f4f5',
      border: `1px solid ${isDark ? '#52525b' : '#a1a1aa'}`,
      borderRadius: 8,
      color: isDark ? '#a1a1aa' : '#52525b',
    }
  }
  if (isStart) {
    return {
      background: isDark ? '#052e16' : '#f0fdf4',
      border: `2px solid ${isDark ? '#16a34a' : '#22c55e'}`,
      borderRadius: 8,
      color: isDark ? '#86efac' : '#166534',
    }
  }
  if (isEnd || isTerminal) {
    return {
      background: isDark ? '#1c1917' : '#fefce8',
      border: `2px solid ${isDark ? '#ca8a04' : '#eab308'}`,
      borderRadius: 8,
      color: isDark ? '#fde047' : '#713f12',
    }
  }
  return {
    background: isDark ? '#18181b' : '#ffffff',
    border: `1px solid ${isDark ? '#3f3f46' : '#e4e4e7'}`,
    borderRadius: 8,
    color: isDark ? '#d4d4d8' : '#18181b',
  }
}

export function WorkflowGraph({ definition, currentStateId, visitedStates = [], onNodeClick }: Props) {
  const isDark = useIsDark()
  const { nodes: graphNodes, edges: graphEdges } = useMemo(
    () => layoutGraph(definition),
    [definition]
  )

  const rfNodes: Node[] = graphNodes.map((n) => ({
    id: n.id,
    position: { x: n.x, y: n.y },
    data: {
      label: (
        <div className="px-3 py-2 text-center">
          <div className="text-sm font-medium">{n.label}</div>
          {n.taskCount > 0 && (
            <div className="text-xs mt-0.5 opacity-70">{n.taskCount} task{n.taskCount > 1 ? 's' : ''}</div>
          )}
          {n.isStart && <div className="text-xs font-medium mt-0.5">START</div>}
          {(n.isEnd || (n.isTerminal && !n.isStart)) && <div className="text-xs font-medium mt-0.5">END</div>}
        </div>
      ),
    },
    style: nodeStyle(
      n.isStart,
      n.isEnd,
      n.isTerminal,
      n.id === currentStateId,
      visitedStates.includes(n.id),
      isDark
    ),
    width: n.width,
    height: n.height,
  }))

  const edgeColor = isDark ? '#52525b' : '#a1a1aa'
  const edgeLabelColor = isDark ? '#a1a1aa' : '#71717a'
  const edgeLabelBg = isDark ? '#18181b' : 'white'

  const rfEdges: Edge[] = graphEdges.map((e) => ({
    id: e.id,
    source: e.from,
    target: e.to,
    label: e.label,
    markerEnd: { type: MarkerType.ArrowClosed, color: edgeColor },
    style: { stroke: edgeColor },
    labelStyle: { fontSize: 11, fill: edgeLabelColor },
    labelBgStyle: { fill: edgeLabelBg, fillOpacity: 0.85 },
  }))

  const handleNodeClick = useCallback(
    (_: unknown, node: Node) => onNodeClick?.(node.id),
    [onNodeClick]
  )

  return (
    <div style={{ width: '100%', height: '100%' }}>
      <ReactFlow
        nodes={rfNodes}
        edges={rfEdges}
        onNodeClick={handleNodeClick}
        fitView
        fitViewOptions={{ padding: 0.2 }}
        nodesDraggable={false}
        nodesConnectable={false}
        elementsSelectable={!!onNodeClick}
      >
        <Background color={isDark ? '#27272a' : '#e4e4e7'} gap={20} />
        <Controls showInteractive={false} />
        <MiniMap
          nodeStrokeWidth={3}
          zoomable
          pannable
          style={{ background: isDark ? '#18181b' : '#f4f4f5' }}
          maskColor={isDark ? 'rgba(0,0,0,0.4)' : 'rgba(240,240,240,0.6)'}
        />
      </ReactFlow>
    </div>
  )
}
