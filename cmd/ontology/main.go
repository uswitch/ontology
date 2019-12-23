package main

import (
	"bufio"
	"context"
	"log"
	"os"

	"github.com/uswitch/ontology/pkg/store/gremlin"

	"github.com/uswitch/ontology/pkg/types"
	_ "github.com/uswitch/ontology/pkg/types/entity"
	_ "github.com/uswitch/ontology/pkg/types/entity/v1"
	_ "github.com/uswitch/ontology/pkg/types/relation"
	_ "github.com/uswitch/ontology/pkg/types/relation/v1"
)

func main() {
	store, err := gremlin.NewLocalServer("ws://127.0.0.1:8182")
	if err != nil {
		log.Fatalf("failed to connect to store: %v", err)
	}

	instances := []types.Instance{}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		in, err := types.Parse(scanner.Text())
		if err != nil {
			log.Printf("error parsing: %v", err)
			continue
		}

		instances = append(instances, in)
		//log.Printf("%T %+v", out, out)
	}
	if err := scanner.Err(); err != nil {
		log.Printf("reading standard input: %v", err)
	}

	if err = store.Add(context.Background(), instances...); err != nil {
		log.Printf("error adding: %v", err)
	}

	/*	out, err := Parse(`
		{
		  "metadata": {
		    "type": "/relation/v1/was_built_by"
		  },
		  "properties": {
		    "a": "asdf",
		    "b": "sdfg",
		    "ref": "dfgh",
		    "at": "fghj"
		  }
		}
		`)
			log.Printf("%v %+v", err, out)

			jout, _ := json.Marshal(out)
			log.Println(string(jout))*/
}
