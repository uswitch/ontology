
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

We could take a few different tacks to enable event subscription:
  - [etcd](https://etcd.io), used by the Kubernetes API server, allows us to support watching
    entities/classes easily
  - DynamoDB/Firestore both support a stream of changes API
  - Kafka could also work, though is the most complicated of the three

~We will not worry about the long term persistence of the store in Ontology as it should be reconsituted
from the static representation of slowly changing entities, and resyncing dynamic changing entities.~
We need to persistent relations betweens entities.

Much like Kubernetes, we should version our classes to allow for evolution of their expected attributes.

The kinds of things we might want to model:
- type
  - data classification
  - partnership

  - person
  - role
  - team

  - asset
    - third party
    - service
    - gdrive document
    - gdrive sheet
    - gdrive slides
    - repo
    - container
    - kubernetes pod
    - kubernetes secret
    - aws instance
    - physical computer
    - gcp project
    - gcp bq table

We will probably want to define some relationships like:
  - (thing) is a (type)
  - (person) works on (team)
  - (person) works as a (role)
  - (team) is a sub-team of (team)
  - (person/team) owns (asset)
  - (person/team) has read access to (asset)
  - (person/team) has write access to (asset)
  - (asset) is part of (asset)
  - (asset) is used by (partnership)
  - (asset) is classified as (data classification)

We will need to be able to extrapolate these relationships to do things like

  - `engineer` is a `role`
  - `jane` is a `person`
  - `energy` is a `team`
  - `energy-wiki` is a `repo`

  - jane works as a engineer
  - jane is part of energy

  - (all engineers in energy) has write access to energy-wiki

`(all engineers in energy)` is a noun that is rule based, rather than list based

Off the basis of the above statements we should be able to list the people that have access to a given repo

Another example of how we would expect to fall upwards for
  - `logging` is a `service`
  - `kibana` is a `service`
  - `restricted` is a `data classification`
  - `operational` is a `data classification`
  - `cloud/kibana-34gg5` is a `kubernetes pod`
  - `declassified-logs` is a `aws s3 bucket`

  - `logging` is classified as `restricted`
  - `kibana` is part of `logging`
  - `cloud/kibana-34gg5` is part of `kibana`
  - `declassified-logs` is part of `logging`
  - `declassified-logs` is classified as `operational`

We should be able to infer that the pod is classified restricted, but the bucket is just operational.


Thinking that it might make sense to use a combo of OPA, Dynamo and Golang to build it.
  - we are planning to use OPA for policy more broadly, it has Datalog as an acestor so should be
    avble to do the above



Learning from examples.rego/data.json:

OPA doesn't support recusion so we have to unroll recursion to a certain depth. We might be able to calculate
the max depth required before generating the OPA. Recursion seems to be only required on relationships with
the same type in and out. we have different directions for sub teams and part of.

We are likely to have name/id collisions in asset compound types where we lookup across a few types
