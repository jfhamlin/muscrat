package graph

import (
	"reflect"
	"testing"

	"github.com/glojurelang/glojure/pkg/glj"
)

func TestAlignGraphs(t *testing.T) {
	type args struct {
		a any
		b any
	}
	tests := []struct {
		name string
		args args
		want GraphAlignment
	}{
		{
			name: "simple",
			args: args{
				a: readGraph(`{
:nodes ({:id "1", :type :sin, :args [], :key nil, :sink nil}
        {:id "2", :type :const, :args [200.0], :key nil, :sink nil}
        {:id "3", :type :out, :ctor nil, :args [0], :key nil, :sink true}
        {:id "4", :type :out, :ctor nil, :args [1], :key nil, :sink true}),
:edges ({:from "2", :to "1", :port "w"}
        {:from "1", :to "3", :port "in"}
        {:from "1", :to "4", :port "in"})
}
`),
				b: readGraph(`{
:nodes ({:id "10", :type :sin, :args [], :key nil, :sink nil}
        {:id "20", :type :const, :args [200.0], :key nil, :sink nil}
        {:id "30", :type :out, :ctor nil, :args [0], :key nil, :sink true}
        {:id "40", :type :out, :ctor nil, :args [1], :key nil, :sink true}),
:edges ({:from "20", :to "10", :port "w"}
        {:from "10", :to "30", :port "in"}
        {:from "10", :to "40", :port "in"})
}
`),
			},
			want: GraphAlignment{
				NodeIdentities: map[NodeID]NodeID{
					"10": "1",
					"20": "2",
					"30": "3",
					"40": "4",
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			a := SExprToGraph(tt.args.a)
			b := SExprToGraph(tt.args.b)
			if got := AlignGraphs(a, b); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AlignGraphs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func readGraph(s string) any {
	return glj.Read(s)
}
