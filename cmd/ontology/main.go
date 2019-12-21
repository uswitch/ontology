package main

import (
	"bufio"
	"log"
	"os"

	"github.com/uswitch/ontology/pkg/types"
	_ "github.com/uswitch/ontology/pkg/types/entity"
	_ "github.com/uswitch/ontology/pkg/types/entity/v1"
	_ "github.com/uswitch/ontology/pkg/types/relation"
	_ "github.com/uswitch/ontology/pkg/types/relation/v1"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		_, err := types.Parse(scanner.Text())
		if err != nil {
			log.Printf("error parsing: %v", err)
		}

		//log.Printf("%T %+v", out, out)
	}
	if err := scanner.Err(); err != nil {
		log.Printf("reading standard input: %v", err)
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
