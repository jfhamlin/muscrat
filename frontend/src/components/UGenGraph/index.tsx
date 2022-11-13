import React, {
  useMemo,
}from 'react';

//@ts-ignore
import Graph from "react-graph-vis";

interface Node {
  id: any;
  label: string;
}

interface Edge {
  from: any;
  to: any;
}

interface Graph {
  nodes: Node[];
  edges: Edge[];
}

interface UGenGraphProps {
  graph: Graph;
}

export default function UGenGraph(props: UGenGraphProps) {
  const options = {
    layout: {
      hierarchical: true
    },
    edges: {
      color: "#000000"
    },
    height: "500px"
  };

  const events = {
    select: function (event: any) {
      console.log('select!', event);
      var { nodes, edges } = event;
    }
  };

  const graphJson = JSON.stringify(props.graph);

  // BUGBUG: this is a hack to get around the fact that react-graph-vis
  // can't handle a changing graph. https://github.com/crubier/react-graph-vis/issues/92
  const graphComponent = useMemo(() => {
    return <Graph graph={JSON.parse(graphJson)} options={options} events={events} />
  }, [graphJson]);

  return (
    <div>
      {graphComponent}
    </div>
  );
}
