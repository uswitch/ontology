
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

```
metadata:
  annotations:
    cloud.rvu.ontology/relation/v1/is_part_of: '["/rvu/mortgages/page-speed", "/rvu/mortgages/bankrate/front-end"]'
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

## Output

Mustache templates used as they do not limit us to Ruby, which is mostly a language to prototype this in.
