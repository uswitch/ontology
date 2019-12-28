
A service for holding a description of our business and the related assets.

Its design influenced by [Ontology](https://en.wikipedia.org/wiki/Ontology_(information_science)),
[Ontology by Jane Street](https://www.janestreet.com/tech-talks/a-language-oriented-system-design/)
and the Kubernetes API Server.

Integration is the name of the game. We need to be able to pull data from authoritive systems (AWS,
Azure, ...) and configure other systems based on the relationships and rules defined here.

We want to keep the impl of this a simple as possible, using no new tools to us. Probably don't want to write our own type system and checker.

Some wants for the system:
  - Slowly changing entities about the structure of the business. For example: teams, partnerships,
    services
  - Dynamically changing entities synced in from other systems. For example: people, computers,
    documents
  - Classes, that entities and relationships will belong to, that enforce a schema on attributes
  - Relationship between entities
  - Two person sign-off for the changes in entities and relations. Except for where synced, we trust
    there has been verification in the upstream system
  - Audit log of all changes to the system
  - Event subscriptions for changes in entites for outside agents to act on

## Structure

This repo will act both as the store for slowly changing entities, as well as containing the code
for aggregating entities from other systems.

We can see these different types of entities/relations as being internal and external to ontology.
We will store all the internal (slowly changing) entities in `./internal/...` and the external (
dynamicly changing) entities in `./external/...`. The former will be managed by humans changing
them via pull requests to this repo, and the later loaded in via `./bin/sync`.

## Labeling

We use labels/tags on entities defined outside of this repository in order to describe
relations.

There are utilities in `./bin` to label resources in the right way.

The value of labels should either be a string, or an array of strings. Sometimes we are
constrained by the schema of the data - looking at you, Kubernetes - so will have to represent
an array using stringified JSON.

### Kubernetes labeling example

Kubernetes is a devil and doesn't allow paths after the subdomain, so we have to use `.` instead of `/`
as a separator.

```
metadata:
  annotations:
    cloud.rvu.ontology/relation.v1.is_part_of: '["/rvu/mortgages/page-speed", "/rvu/mortgages/bankrate/front-end"]'
```

### Github labeling example

A file `/.github/metadata` should exist in the repository and look like the following:

```
metadata:
  labels:
    cloud.rvu.ontology/relation/v1/is_part_of:
      - /rvu/airship/observability/logging
      - /rvu/airship/provisioning
```

### AWS labeling example

Similar to Kubernetes we will need to encode JSON in the value in order to add multiple of the same kind
of relation.

TODO: We should and AWS Config rememdiation for every account we make to default the value of is_part_of to
the team that owns the account. Maybe not, this might cause some issues when terraform runs and tries to remove the tags. Yay for a race between the two!


### Docker container labeling example

We look for some of the standard docker labels for which repository/commit it was built for and convert
them into relations. Luckily Drone has been adding these labels onto containers for us.

## Types

We, unsuprisingly, have the notion of types of things, be it a relation or an entity. The type system is a
simple hierarchical system, with single parents. This allows us to create some constraints around what
things are related and expected properties of relations and entities.

We express the constraints on properties using JSON schema.

We've added one extension to JSON Schema, which is the `pointer_to` keyword. This is used in relations
to constrain the entities that can be related. This relies on the type hierarchy.

These types aren't defined externally of Ontology.

### Internal entities

We need a type to define what is expected of a type

  - `/type`
    A `/type` requires the entity to define a spec property which contains the constraints on properties
    of any instantiations of the type. Optionally, you can also define a `parent` property that lets you
    build up a hierarchy.

### Internal relations [Currently unimplemnted]

These are required to map the YAML metadata into relations.

  - `/relation/v1/is_a` -> `.metadata.type`
  - `/relation/v1/is_named` -> `.metedata.name`
  - `/relation/v1/was_updated_at` -> `.metadata.update_at`

### Required metadata

  - `type`: References an entity that is derived from `/type`
  - `id`: Any unique string identifier
  - `name`: A more human identifier for the entity
  - `updated_at`[Current implemented]: An ISO8601 formated string that contains when the information
    contained in the entity is from.

## Ingesting data

As set out in the requirements Ontology revolves around a combination of static, slow moving entities and
dynamic, more quickly changing entities.

We support loading entities and relations from the file system in two formats: JSON files organised in a
folder heirarchy, or as files containing many JSON objects separated by new lines. These two formats are
supported when initially starting up the application and are used to preload the types, static entities and
potentially, initial snapshots of dynamic entities.

Once Ontology is started it will expose an HTTP API to allow the creation and mutation of entities.

The truth of Ontology is a product of the data put into it. It is important to ensure that only the systems
that are expect to be mutating entities and relations are the ones that do. If we ever want to use it to
make security decisions this is important.

**Authnz + Auditing**. This could be combined with something like the Certificate Transparency backing store.
It provides a tamper proof log of things it's been told

### Sync agents

We've written a set of agents to aggregate data from external systems:
  - AWS,
  - GitHub,
  - Docker Registry v2,
  - Kubernetes,
  - Azure AD.

### Formats

We started with a single format, but this might be splintering into the two views of the same data.

We wanted schema for the data that is stored against entities so we have the concept of types. These build up
a JSON schema object that can be used to validate an entity.

#### YAML

I can't think of the right name for this. It is the complete definition of an entity or relation in one blob.
This looks a lot like a Kubernetes resource, as this has influenced my thinking a bunch of late.

```
metadata:
  type: /entity/v1/thing
  id: /things/wibble
properties:
  thinginess: 7
```

The only required fields a `.metadata.type` and `.metadata.id`. The fields `.metadata.updated_at` and
`.metadata.name` will always be available, but if not provided will be derived by Ontology.


#### EAV (Entity-attribute-value) [Current unimplemented]

Break up the data into triples.

This might end up being mostly an internal format. This is how Cayley et al store there graph data.

```
( id             , attribute                   , value )
( /things/wibble , /relation/v1/is_a           , /entity/v1/thing )
( /things/wibble , /entity/v1/thing#thinginess , 7 )
```

The `( id , is_a , type)` would be the only required triple, as in the YAML format above. There
might be other expected triples based on the type of the entity.

Properties of both entities and relations will be represented by the type followed by `#` and then
the path of the property. In the above example, `thinginess` is a property of `/entity/v1/thing`.

## Output

Mustache templates used as they do not limit us to Ruby, which is mostly a language to prototype this in.





## JanusGraph (gremlin server)

Here are some notes on how to run a local JanusGraph server for use when developing
Ontology

### Spinning up a server

```
$ mkdir -p var/janusgraph      # .gitignored
$ docker run --rm --env-file etc/janusgraph/env.list -p 8182:8182 -v $(pwd)/var/janusgraph:/var/lib/janusgraph -v $(pwd)/etc/janusgraph:/etc/opt/janusgraph:ro --name janusgraph janusgraph/janusgraph:0.4
```

### Test server

Different port and var directory

```
$ docker run --rm --env-file etc/janusgraph/env.list -p 8183:8182 -v $(pwd)/var/janusgraph-test:/var/lib/janusgraph -v $(pwd)/etc/janusgraph:/etc/opt/janusgraph:ro --name janusgraph-test janusgraph/janusgraph:0.4
```

### Start Gremlin console

```
$ docker run --rm --link janusgraph:janusgraph -e GREMLIN_REMOTE_HOSTS=janusgraph -it janusgraph/janusgraph:0.4 ./bin/gremlin.sh
```

### Configure, enable and check status of Indicies

```
gremlin> :remote connect tinkerpop.server conf/remote.yaml


gremlin> :> mgmt = graph.openManagement(); mgmt.buildIndex('vertexByID', Vertex.class).addKey(mgmt.getOrCreatePropertyKey('id')).buildCompositeIndex(); mgmt.commit()
gremlin> :> mgmt = graph.openManagement(); mgmt.buildIndex('edgeByID', Edge.class).addKey(mgmt.getOrCreatePropertyKey('id')).buildCompositeIndex(); mgmt.commit()
gremlin> :> mgmt = graph.openManagement(); mgmt.buildIndex('vertexByType', Vertex.class).addKey(mgmt.getOrCreatePropertyKey('type')).buildCompositeIndex(); mgmt.commit()
gremlin> :> mgmt = graph.openManagement(); mgmt.buildIndex('edgeByType', Edge.class).addKey(mgmt.getOrCreatePropertyKey('type')).buildCompositeIndex(); mgmt.commit()

gremlin> :> m = graph.openManagement(); index = m.getGraphIndex("edgeByID"); m.updateIndex(index, SchemaAction.ENABLE_INDEX).get(); m.commit()
gremlin> :> m = graph.openManagement(); index = m.getGraphIndex("edgeByType"); m.updateIndex(index, SchemaAction.ENABLE_INDEX).get(); m.commit()


gremlin> :> m = graph.openManagement(); index = m.getGraphIndex("byID"); pkey = index.getFieldKeys()[0]; index.getIndexStatus(pkey)
gremlin> :> g.E().drop().iterate(); g.V().drop().iterate()
gremlin> :> graph.getOpenTransactions()
gremlin> :> graph.getOpenTransactions().getAt(0).rollback()
```
