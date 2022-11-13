import React, {
  useMemo,
  useCallback,
  useEffect,
}from 'react';

import ReactFlow, {
  MiniMap,
  Controls,
  Background,
  useNodesState,
  useEdgesState,
  addEdge,
} from 'reactflow';

import { forceSimulation } from 'd3-force';
import * as d3 from 'd3';

interface Node {
  id: string;
  position: {
    x: number;
    y: number;
  };
  data: {
    label: string;
  };
}

interface Edge {
  id: string;
  source: any;
  target: any;
}

interface Graph {
  nodes: Node[];
  edges: Edge[];
}

interface UGenGraphProps {
  graph: Graph;
}

import 'reactflow/dist/style.css';

export default function UGenGraph(props: UGenGraphProps) {
  const inNodes = props.graph.nodes;
  const edges = props.graph.edges;

  const placedNodes = useMemo(() => layoutNodes(inNodes, edges), [inNodes, edges]);
  const [nodes, setNodes, onNodesChange] = useNodesState(placedNodes);
  useEffect(() => {
    setNodes(placedNodes);
  }, [placedNodes, setNodes]);

  /* const [edges, setEdges, onEdgesChange] = useEdgesState(props.initialGraph.edges); */

  /* const [nodes, setNodes, onNodesChange] = useNodesState(NODES.slice(0, 2));
   * const [edges, setEdges, onEdgesChange] = useEdgesState([]); */

  /* console.log('nodes', nodes);
   * console.log('edges', edges); */

  // do nothing
  const onConnect = useCallback((params) => {})//setEdges((eds) => addEdge(params, eds)), [setEdges]);

  {/* onEdgesChange={onEdgesChange}
      onConnect={onConnect} */}
  return (
    <ReactFlow
      nodes={nodes}
      edges={edges}
      onNodesChange={onNodesChange}
      fitView
    >
      <MiniMap />
      <Controls />
      <Background />
    </ReactFlow>
  );
}

function layoutNodes(nodes, edges) {
  const simNodes = nodes.map((d, i) => ({ id: d.id, x: 0, y: 0 }));
  const simEdges = edges.map((d) => ({ source: d.source, target: d.target }));

  const nodeWidth = 200;
  const nodeHeight = 50;

  const simNodesById = simNodes.reduce((acc, d) => {
    acc[d.id] = d;
    return acc;
  }, {});

  const edgesBySource = edges.reduce((acc, d) => {
    if (!acc[d.source]) {
      acc[d.source] = [];
    }
    acc[d.source].push(d);
    return acc;
  }, {});

  const edgesByTarget = edges.reduce((acc, d) => {
    if (!acc[d.target]) {
      acc[d.target] = [];
    }
    acc[d.target].push(d);
    return acc;
  }, {});

  // dumb position initialization, iterating over all edges multiple
  // times and placing source nodes above target nodes. note that y increases
  // downwards in SVG coordinates.
  for (let i = 0; i < 100; ++i) {
    simEdges.forEach((d) => {
      const source = simNodesById[d.source];
      const target = simNodesById[d.target];
      if (source.y < target.y - 100) {
        return;
      }
      source.y = 0.5 * (source.y + target.y - 100);
    });
  }

  // a node's target x position is the average of its source and
  // target x positions
  const nodeTargetX = (id: string) => {
    const sourceEdges = edgesBySource[id] ?? [];
    const targetEdges = edgesByTarget[id] ?? [];
    const sourceX = sourceEdges.reduce((acc, d) => acc + simNodesById[d.target].x, 0) / sourceEdges.length;
    const targetX = targetEdges.reduce((acc, d) => acc + simNodesById[d.source].x, 0) / targetEdges.length;
    return 0.5 * (sourceX + targetX);
  };

  // spread out nodes at nearby y values horizontally
  for (let i = 0; i < 1; ++i) {
    simNodes.forEach((d) => {
      const nearby = simNodes.filter((d2) => Math.abs(d2.y - d.y) < nodeHeight);
      const sorted = nearby.sort((a, b) => nodeTargetX(a.id) - nodeTargetX(b.id));
      const sortedTargetX = sorted.map((d2) => nodeTargetX(d2.id));
      console.log('sortedTargetX', sortedTargetX);
      // distribute nodes near their target x position, but avoid collisions
      for (let j = 0; j < 10; ++j) {
        sorted.forEach((d2, i) => {
          if (i == sorted.length - 1) {
            return;
          }
          if (sortedTargetX[i + 1] - sortedTargetX[i] < nodeWidth) {
            sortedTargetX[i] = 0.5 * (sortedTargetX[i] + sortedTargetX[i + 1] - nodeWidth);
          }
        })
      }
      sorted.forEach((d2, i) => {
        //d2.x = sortedTargetX[i];
      });

      // distribute nodes evenly along the x axis around the center
      // sort by target x position

      /* 
       * const rowWidth = nearby.length * nodeWidth; */
      /* sorted.forEach((d2, i) => {
       *   d2.x = (centerX - 0.5 * rowWidth) + i * nodeWidth;
       * }); */
      //d.x = nodeTargetX(d.id);
    });
  }

  /* const forceNode = d3.forceManyBody();
   * const forceLink = d3.forceLink(simEdges).id((d) => d.id); */

  /* const simulation = forceSimulation(simNodes)
   *   //.force('link', forceLink)
   *   .force('charge', d3.forceManyBody())
   *   .force('center', d3.forceCenter(nodeWidth / 2, nodeHeight / 2))
   *   .force('collide', d3.forceCollide(10));
   * simulation.stop();


   * for (let i = 0; i < 1000; ++i) simulation.tick(); */

  return nodes.map((node, i) => {
    return {
      position: {
        x: simNodesById[node.id].x,
        y: simNodesById[node.id].y,
      },
      ...node,
    };
  });
}
