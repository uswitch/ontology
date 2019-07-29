require 'digest'

def clone_hash(hash)
  Marshal.load(Marshal.dump(hash))
end

def alias_entity(entity, id:, aliases:)
  entity_ids = [id] + aliases

  entities = entity_ids.map { |entity_id|
    e = clone_hash(entity)
    e[:metadata][:id] = entity_id
    e
  }

  relations = aliases.map { |alias_id|
    {
      metadata: {
        id: "#{id}/#{Digest::SHA256.hexdigest(alias_id)}",
        type: "/relation/v1/is_the_same_as",
      },
      properties: {
        a: alias_id,
        b: id,
      },
    }
  }

  entities + relations
end

def add_ids_to(things, base:)
  things.each_with_index.map { |thing, i|
    thing[:metadata][:id] = "#{base}/#{i}"
    thing
  }
end
