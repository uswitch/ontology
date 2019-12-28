package main

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"github.com/uswitch/ontology/pkg/store"
	"github.com/uswitch/ontology/pkg/store/gremlin"
	"github.com/uswitch/ontology/pkg/types"
	_ "github.com/uswitch/ontology/pkg/types/entity"
	_ "github.com/uswitch/ontology/pkg/types/entity/v1"
	"github.com/uswitch/ontology/pkg/types/relation"
	_ "github.com/uswitch/ontology/pkg/types/relation/v1"
)

var (
	rootID            string
	searchDepth       int
	traverseDirection string
	constrain         []string
)

var dotCmd = &cobra.Command{
	Use:   "dot",
	Short: "Output graph in dot format",
	Args:  cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		s, err := gremlin.NewLocalServer(serverURL)
		if err != nil {
			log.Fatalf("failed to connect to store: %v", err)
		}

		typesToDot := []types.ID{relation.ID}

		if len(args) > 0 {
			typesToDot = make([]types.ID, len(args))
			for idx, arg := range args {
				typesToDot[idx] = types.ID(arg)
			}
		}

		allInsts := []types.Instance{}

		if rootID == "" {
			if allInsts, err = s.ListByType(context.Background(), typesToDot, store.ListByTypeOptions{IncludeSubclasses: true}); err != nil {
				log.Fatalf("error getting relations: %v", err)
			}
		} else {
			constrainIDs := make([]types.ID, len(constrain))

			for idx, str := range constrain {
				constrainIDs[idx] = types.ID(str)
			}

			if allInsts, err = s.ListFromByType(
				context.Background(), types.ID(rootID), typesToDot,
				store.ListFromByTypeOptions{
					ListByTypeOptions: store.ListByTypeOptions{IncludeSubclasses: true},

					MaxDepth:        searchDepth,
					Direction:       store.TraverseDirectionFrom(traverseDirection),
					ConstrainByType: constrainIDs,
				}); err != nil {
				log.Fatalf("error getting relations: %v", err)
			}
		}

		fmt.Println("digraph {")

		for _, inst := range allInsts {
			relation := inst.(relation.Instance)
			fmt.Printf("\"%s\" -> \"%s\";\n", relation.A(), relation.B())
		}

		fmt.Println("}")

	},
}

func init() {
	dotCmd.Flags().StringVar(
		&rootID, "root-id",
		"", "id of a vertex to start with",
	)
	dotCmd.Flags().IntVar(&searchDepth, "search-depth", 2, "How many relations to traverse before we stop")
	dotCmd.Flags().StringVar(
		&traverseDirection, "traverse-direction",
		"out", "What direction to traverse the graph when rooting on an entity",
	)
	dotCmd.Flags().StringSliceVar(
		&constrain, "constrain",
		[]string{}, "Constraint the relations that will be traversed to only these types",
	)

	rootCmd.AddCommand(dotCmd)
}
